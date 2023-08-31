package sql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/h3ll0kitt1/avitotest/internal/config"
	"github.com/h3ll0kitt1/avitotest/internal/models"
)

type SQLStorage struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewStorage(cfg config.Database, logger *zap.SugaredLogger) (*SQLStorage, error) {

	DSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DATABASE_HOST, cfg.POSTGRES_USER, cfg.POSTGRES_PASSWORD, cfg.POSTGRES_DB)

	db, err := sql.Open("pgx", DSN)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ctx := context.Background()
	query := `CREATE TABLE IF NOT EXISTS users(
		id integer primary key)`
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return nil, err
	}

	query = `CREATE TABLE IF NOT EXISTS segments(
		slug varchar(255) primary key)`
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return nil, err
	}

	query = `CREATE TABLE IF NOT EXISTS users_segments(
		user_id integer references users (id) on delete cascade not null,
		segment_slug varchar(255) references segments (slug) on delete cascade not null,
		expires_at timestamp,
		unique (user_id, segment_slug))`
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return nil, err
	}

	query = `CREATE TABLE IF NOT EXISTS segments_history(
		user_id integer not null,
		segment_slug varchar(255) not null,
		action boolean not null,
		action_time TIMESTAMP not null)`
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return nil, err
	}
	tx.Commit()

	return &SQLStorage{
		db:     db,
		logger: logger,
	}, nil
}

func (s *SQLStorage) CreateSegment(ctx context.Context, slug string, PercentageRND int) error {

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Добавляем сегмент, если его не существует
	query := ` INSERT INTO segments (slug) VALUES ($1) ON CONFLICT (slug) DO NOTHING`
	_, err = tx.ExecContext(ctx, query, slug)
	if err != nil {
		return err
	}

	// Если было передано значение желаемого процента случайных пользователей
	if PercentageRND != 0 {

		// Выбираем случайных пользователей
		usersRND, err := s.getRandomUsers(ctx, PercentageRND)
		if err != nil {
			return err
		}
		s.logger.Infow("info",
			"CreateSegment: users chosen at random: ", usersRND,
		)

		for _, user := range usersRND {

			// Добавляем сегмент случайному пользователю
			query := ` 	INSERT INTO users_segments (user_id, segment_slug, expires_at) 
						VALUES ($1, $2, null) 
						ON CONFLICT (user_id, segment_slug) DO NOTHING`
			_, err := tx.ExecContext(ctx, query, user, slug)
			if err != nil {
				return err
			}

			// Добавляем запись о добавлении в историю
			query = ` 	INSERT INTO segments_history (user_id, segment_slug, action, action_time)
    					VALUES ($1, $2, true, now())`
			_, err = tx.ExecContext(ctx, query, user, slug)
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (s *SQLStorage) DeleteSegment(ctx context.Context, slug string) error {

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Получаем список пользователей для которых необходимо удалить сегмент
	users, err := s.getUsersInSegment(ctx, slug)
	if err != nil {
		return err
	}
	s.logger.Infow("info",
		"DeleteSegment: users currently in segment: ", users,
	)

	// Для каждого пользователя из списка вносим в историю информацию об удалении
	for _, user := range users {

		query := ` 	INSERT INTO segments_history (user_id, segment_slug, action, action_time)
					VALUES ($1, $2, false,  now())`
		_, err := tx.ExecContext(ctx, query, user, slug)
		if err != nil {
			return err
		}
	}

	// Удаляем сегмент из таблицы сегментов
	query := `	DELETE FROM segments
				WHERE slug = $1`

	_, err = tx.ExecContext(ctx, query, slug)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLStorage) GetSegmentsByUserID(ctx context.Context, user int64) ([]models.Segment, error) {

	segments := make([]models.Segment, 0)

	query := `	SELECT segment_slug FROM users_segments
				WHERE user_id = $1 AND (expires_at >= NOW() OR expires_at IS NULL)`
	rows, err := s.db.QueryContext(ctx, query, user)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var segment models.Segment
		err = rows.Scan(&segment.Slug)
		if err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	s.logger.Infow("info",
		"GetSegmentsByUserID: user is currently in segments: ", segments,
	)
	return segments, nil
}

func (s *SQLStorage) UpdateSegmentsByUserID(ctx context.Context, user int64, deleteList []models.Segment, addList []models.Segment) error {

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Добавляем пользователя, если его не существует
	query := ` 	INSERT INTO users (id) VALUES ($1)
     			ON CONFLICT (id) DO NOTHING`

	_, err = tx.ExecContext(ctx, query, user)
	if err != nil {
		return err
	}

	// Удаляем сегмент, если пользователь находится в нем и вносим удаление в историю
	for _, segment := range deleteList {

		ok, err := s.checkUserInSegment(ctx, user, segment.Slug)
		if err != nil {
			return err
		}

		if !ok {
			continue
		}

		query = ` 	DELETE FROM users_segments
   					WHERE user_id = $1 AND segment_slug = $2`
		_, err = tx.ExecContext(ctx, query, user, segment.Slug)
		if err != nil {
			return err
		}

		query = ` 	INSERT INTO segments_history (user_id, segment_slug, action, action_time)
    				VALUES ($1, $2, false, now())`

		_, err = tx.ExecContext(ctx, query, user, segment.Slug)
		if err != nil {
			return err
		}
	}

	for _, segment := range addList {

		//Добавляем новые сегменты
		query := ` 	INSERT INTO segments (slug) VALUES ($1)
     			    ON CONFLICT (slug) DO NOTHING`

		_, err = tx.ExecContext(ctx, query, segment.Slug)
		if err != nil {
			return err
		}

		// Если указан TTL, тогда вычисляем время, когда сегмент должен перестать быть валидным и обновляем в expires_at
		if segment.DaysTTL != 0 {
			query = ` 	INSERT INTO users_segments (user_id, segment_slug, expires_at)
						VALUES ($1, $2, now() + interval '1 day' * $3)
						ON CONFLICT (user_id, segment_slug) DO UPDATE
						SET expires_at = EXCLUDED.expires_at`
			_, err = tx.ExecContext(ctx, query, user, segment.Slug, segment.DaysTTL)
			if err != nil {
				return err
			}
			// Иначе считаем, что пользователя необходимо добавить в сегмент перманентно (обозначается NULL)
		} else {
			query = ` 	INSERT INTO users_segments (user_id, segment_slug, expires_at)
						VALUES ($1, $2, null)
						ON CONFLICT (user_id, segment_slug) DO UPDATE
						SET expires_at = EXCLUDED.expires_at`
			_, err = tx.ExecContext(ctx, query, user, segment.Slug)
			if err != nil {
				return err
			}
		}

		// Пишем о добавлении в историю
		query = ` 	INSERT INTO segments_history (user_id, segment_slug, action, action_time)
    				VALUES ($1, $2, true, now())`
		_, err = tx.ExecContext(ctx, query, user, segment.Slug)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLStorage) GetHistory(ctx context.Context, users []int64, days int) ([]models.History, error) {

	usersHistory := make([]models.History, 0)

	for _, user := range users {
		query := `	SELECT segment_slug, user_id, action, action_time 
					FROM segments_history
					WHERE user_id = $1 AND action_time >= NOW() - interval '1 day' * $2;`

		rows, err := s.db.QueryContext(ctx, query, user, days)

		for rows.Next() {
			var history models.History
			err = rows.Scan(&history.Segment.Slug, &history.User, &history.Action, &history.ActionTime)
			if err != nil {
				return nil, err
			}
			usersHistory = append(usersHistory, history)
		}
		err = rows.Err()
		if err != nil {
			return nil, err
		}

	}
	return usersHistory, nil
}

type expiredSegment struct {
	user int64
	slug string
}

func (s *SQLStorage) DeleteExpiredSegments() {
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Errorw("error",
			"DeleteExpiredSegments: transaction failed: ", err,
		)
		return
	}
	defer tx.Rollback()

	// Найдем все не валидные более для пользователей сегменты
	expiredSegments := make([]expiredSegment, 0)

	query := `	SELECT user_id, segment_slug FROM users_segments
				WHERE  expires_at < now()`
	rows, err := s.db.QueryContext(context.Background(), query)
	if err != nil {
		s.logger.Errorw("error",
			"DeleteExpiredSegments: selection from users_segments failed ", err,
		)
		return
	}

	for rows.Next() {
		var segment expiredSegment
		err = rows.Scan(&segment.user, &segment.slug)
		if err != nil {
			s.logger.Errorw("error",
				"DeleteExpiredSegments: enumerating users_segments failed ", err,
			)
			return
		}
		expiredSegments = append(expiredSegments, segment)
	}
	err = rows.Err()
	if err != nil {
		s.logger.Errorw("error",
			"DeleteExpiredSegments: enumerating users_segments err failed ", err,
		)
		return
	}

	// Пишем об удалении сегмента в историю и удаляем
	for _, segment := range expiredSegments {
		query = ` 	INSERT INTO segments_history (user_id, segment_slug, action, action_time)
    				VALUES ($1, $2, false, now())`
		_, err = tx.ExecContext(context.Background(), query, segment.user, segment.slug)
		if err != nil {
			s.logger.Errorw("error",
				"DeleteExpiredSegments: inserting into segments_history failed ", err,
			)
			return
		}

		query = ` 	DELETE FROM users_segments
   					WHERE user_id = $1 AND segment_slug = $2`
		_, err = tx.ExecContext(context.Background(), query, segment.user, segment.slug)
		if err != nil {
			s.logger.Errorw("error",
				"DeleteExpiredSegments: deleting from segments_history failed ", err,
			)
		}
	}
	s.logger.Infow("info",
		"DeleteExpiredSegments: successfully deleted segments: ", expiredSegments,
	)
	tx.Commit()
}

func (s *SQLStorage) getRandomUsers(ctx context.Context, percentage int) ([]int64, error) {

	usersRND := make([]int64, 0)

	query := `SELECT id FROM users TABLESAMPLE BERNOULLI ($1)`
	rows, err := s.db.QueryContext(ctx, query, percentage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user int64
		err = rows.Scan(&user)
		if err != nil {
			return nil, err
		}
		usersRND = append(usersRND, user)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return usersRND, nil
}

func (s *SQLStorage) getUsersInSegment(ctx context.Context, slug string) ([]int64, error) {

	users := make([]int64, 0)

	query := `SELECT user_id FROM users_segments WHERE segment_slug = $1`
	rows, err := s.db.QueryContext(ctx, query, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user int64
		err = rows.Scan(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *SQLStorage) checkUserInSegment(ctx context.Context, user int64, slug string) (bool, error) {

	query := ` 	SELECT user_id FROM users_segments
   				WHERE user_id = $1 AND segment_slug = $2`

	var userID int
	err := s.db.QueryRowContext(ctx, query, user, slug).Scan(&userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
