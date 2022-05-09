package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
	_ "github.com/lib/pq" // Package pq is a pure Go Postgres driver for the database/sql package
)

const repositorySize = 10

// jobRepositorySQL keeps data in SQL database
type jobRepositorySQL struct {
	Rand      io.Reader
	clock     repository.Clock
	db        *sql.DB
	tableName string
	fields    []string
}

// OpenDB return new database handle
func OpenDB(DBConnString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", DBConnString)
	if err != nil {
		return nil, fmt.Errorf("error connecting to DB: %v", err)
	}

	return db, nil
}

// NewJobRepositorySQL returns new initialized job repository that keeps data in SQL database.
// rand is random number generator, which implements io.Reader. Calling it with nil sets the random number generator
// to the default generator. See repository.GenerateUUID for details.
func NewJobRepositorySQL(clock repository.Clock, db *sql.DB, rand io.Reader) (repository.JobRepository, error) {
	tableName := "jobs"

	if _, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS " + tableName + " (" +
			"uuid UUID PRIMARY KEY, " +
			"created_at VARCHAR(30)," +
			"final_status TEXT, " +
			"channels_download_started_at VARCHAR(30), " +
			"channels_download_finished_at VARCHAR(30), " +
			"users_download_started_at VARCHAR(30), " +
			"users_download_finished_at VARCHAR(30), " +
			"tickets_download_started_at VARCHAR(30), " +
			"tickets_download_finished_at VARCHAR(30), " +
			"excel_files_generation_started_at VARCHAR(30), " +
			"excel_files_generation_finished_at VARCHAR(30), " +
			"emails_sending_started_at VARCHAR(30), " +
			"emails_sending_finished_at VARCHAR(30) " +
			")",
	); err != nil {
		return nil, fmt.Errorf("error creating table %s: %v", tableName, err)
	}

	return &jobRepositorySQL{
		Rand:      rand,
		clock:     clock,
		db:        db,
		tableName: tableName,
		fields: []string{
			"uuid", "created_at", "final_status",
			"channels_download_started_at", "channels_download_finished_at",
			"users_download_started_at", "users_download_finished_at",
			"tickets_download_started_at", "tickets_download_finished_at",
			"excel_files_generation_started_at", "excel_files_generation_finished_at",
			"emails_sending_started_at", "emails_sending_finished_at",
		},
	}, nil
}

func (r jobRepositorySQL) AddJob(ctx context.Context, job job.Job) (ref.UUID, error) {
	jobID, err := repository.GenerateUUID(r.Rand)
	if err != nil {
		return jobID, err
	}

	now := r.clock.NowFormatted().String()

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO "+r.tableName+" ("+r.tableFields()+") VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
		jobID,
		now,
		job.FinalStatus,
		job.ChannelsDownloadStartedAt,
		job.ChannelsDownloadFinishedAt,
		job.UsersDownloadStartedAt,
		job.UsersDownloadFinishedAt,
		job.TicketsDownloadStartedAt,
		job.TicketsDownloadFinishedAt,
		job.ExcelFilesGenerationStartedAt,
		job.ExcelFilesGenerationFinishedAt,
		job.EmailsSendingStartedAt,
		job.EmailsSendingFinishedAt,
	)
	if err != nil {
		return jobID, err
	}

	return jobID, nil
}

func (r jobRepositorySQL) UpdateJob(ctx context.Context, job job.Job) (ref.UUID, error) {
	jobID := job.UUID()

	updateStr := "final_status = $2," +
		"channels_download_started_at = $3, channels_download_finished_at = $4, " +
		"users_download_started_at = $5, users_download_finished_at = $6, " +
		"tickets_download_started_at = $7, tickets_download_finished_at = $8, " +
		"excel_files_generation_started_at = $9, excel_files_generation_finished_at = $10, " +
		"emails_sending_started_at = $11, emails_sending_finished_at = $12"

	stmt, err := r.db.PrepareContext(ctx, "UPDATE "+r.tableName+" SET "+updateStr+" WHERE uuid = $1")
	if err != nil {
		return jobID, err
	}

	_, err = stmt.Exec(
		jobID,
		job.FinalStatus,
		job.ChannelsDownloadStartedAt,
		job.ChannelsDownloadFinishedAt,
		job.UsersDownloadStartedAt,
		job.UsersDownloadFinishedAt,
		job.TicketsDownloadStartedAt,
		job.TicketsDownloadFinishedAt,
		job.ExcelFilesGenerationStartedAt,
		job.ExcelFilesGenerationFinishedAt,
		job.EmailsSendingStartedAt,
		job.EmailsSendingFinishedAt,
	)
	if err != nil {
		return jobID, err
	}

	return jobID, nil
}

func (r jobRepositorySQL) GetJob(ctx context.Context, ID ref.UUID) (job.Job, error) {
	var j job.Job
	var uuid ref.UUID

	if err := r.db.QueryRowContext(ctx, "SELECT "+r.tableFields()+" FROM "+r.tableName+" WHERE uuid = $1", ID).Scan(
		&uuid,
		&j.CreatedAt,
		&j.FinalStatus,
		&j.ChannelsDownloadStartedAt,
		&j.ChannelsDownloadFinishedAt,
		&j.UsersDownloadStartedAt,
		&j.UsersDownloadFinishedAt,
		&j.TicketsDownloadStartedAt,
		&j.TicketsDownloadFinishedAt,
		&j.ExcelFilesGenerationStartedAt,
		&j.ExcelFilesGenerationFinishedAt,
		&j.EmailsSendingStartedAt,
		&j.EmailsSendingFinishedAt,
	); err != nil {
		notFound := errors.Is(err, sql.ErrNoRows)
		if notFound {
			return j, domain.WrapErrorf(repository.ErrNotFound, domain.ErrorCodeNotFound, "error loading job from repository")
		}
		// Something else went wrong!
		return j, err
	}

	if err := j.SetUUID(uuid); err != nil {
		return j, err
	}

	return j, nil
}

func (r jobRepositorySQL) GetLastJob(ctx context.Context) (job.Job, error) {
	var j job.Job
	var uuid ref.UUID

	if err := r.db.QueryRowContext(ctx,
		"SELECT "+r.tableFields()+" FROM "+
			r.tableName+" ORDER BY created_at DESC LIMIT 1").Scan(
		&uuid,
		&j.CreatedAt,
		&j.FinalStatus,
		&j.ChannelsDownloadStartedAt,
		&j.ChannelsDownloadFinishedAt,
		&j.UsersDownloadStartedAt,
		&j.UsersDownloadFinishedAt,
		&j.TicketsDownloadStartedAt,
		&j.TicketsDownloadFinishedAt,
		&j.ExcelFilesGenerationStartedAt,
		&j.ExcelFilesGenerationFinishedAt,
		&j.EmailsSendingStartedAt,
		&j.EmailsSendingFinishedAt,
	); err != nil {
		notFound := errors.Is(err, sql.ErrNoRows)
		if notFound {
			return job.Job{}, domain.NewErrorf(domain.ErrorCodeUnknown, "no jobs in queue")
		}
		// Something else went wrong!
		return j, err
	}

	if err := j.SetUUID(uuid); err != nil {
		return j, err
	}

	return j, nil
}

func (r jobRepositorySQL) ListJobs(ctx context.Context) ([]job.Job, error) {
	// TODO - add pagination?

	var list []job.Job

	rows, err := r.db.QueryContext(
		ctx,
		"SELECT "+r.tableFields()+" FROM "+r.tableName+" ORDER BY created_at DESC LIMIT $1", repositorySize,
	)
	if err != nil {
		return list, err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var j job.Job
		var uuid ref.UUID

		if err := rows.Scan(
			&uuid,
			&j.CreatedAt,
			&j.FinalStatus,
			&j.ChannelsDownloadStartedAt,
			&j.ChannelsDownloadFinishedAt,
			&j.UsersDownloadStartedAt,
			&j.UsersDownloadFinishedAt,
			&j.TicketsDownloadStartedAt,
			&j.TicketsDownloadFinishedAt,
			&j.ExcelFilesGenerationStartedAt,
			&j.ExcelFilesGenerationFinishedAt,
			&j.EmailsSendingStartedAt,
			&j.EmailsSendingFinishedAt,
		); err != nil {
			return list, err
		}

		if err := j.SetUUID(uuid); err != nil {
			return list, err
		}

		list = append(list, j)
	}
	if err := rows.Err(); err != nil {
		return list, err
	}

	return list, nil
}

func (r jobRepositorySQL) tableFields() string {
	return strings.Join(r.fields, ", ")
}
