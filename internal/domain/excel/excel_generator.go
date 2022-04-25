package excel

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type Generator interface {
	// GenerateExcelFiles creates Excel spreadsheet files with tickets' info
	GenerateExcelFiles(ctx context.Context) error

	// DirName returns name of the directory where the files are generated to
	DirName() string
}

func NewExcelGenerator(logger *zap.SugaredLogger, ticketRepository repository.TicketRepository) Generator {
	return &excelGen{
		logger:           logger,
		ticketRepository: ticketRepository,
		dirName:          filepath.Join(os.TempDir(), "reporting-xls-files"),
	}
}

type excelGen struct {
	logger           *zap.SugaredLogger
	ticketRepository repository.TicketRepository
	dirName          string // directory to put generated files to
}

func (g excelGen) DirName() string {
	return g.dirName
}

func (g excelGen) GenerateExcelFiles(ctx context.Context) error {
	emails, err := g.ticketRepository.GetDistinctEmails(ctx)
	if err != nil {
		return err
	}

	if err = os.RemoveAll(g.dirName); err != nil {
		return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not remove directory '%s'", g.dirName)
	}

	err = os.Mkdir(g.dirName, 0750)
	if err != nil {
		return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not create directory '%s'", g.dirName)
	}

	if err := os.Chdir(g.dirName); err != nil {
		return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not change to directory '%s'", g.dirName)
	}

	for _, email := range emails {
		userTickets, err := g.ticketRepository.GetTicketsByEmail(ctx, email)
		if err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not get tickets for email '%s' from repository", email)
		}

		filename := email + ".xlsx"

		f := excelize.NewFile()

		// Set columns width
		sheet := "Sheet1"
		if err := f.SetColWidth(sheet, "A", "B", 13); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetColWidth(sheet, "C", "D", 45); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}

		// Excel file header
		if err := f.SetCellValue(sheet, "A1", "Open tickets assigned to "+email); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "A3", "Ticket type"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "B3", "Number"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "C3", "Short description"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "D3", "Location"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}

		var channelName string
		var row string
		i := 3
		// Excel rows with ticket data
		for _, t := range userTickets {
			i++
			if channelName != t.ChannelName {
				i++
				row = strconv.Itoa(i)
				if err := f.SetCellValue(sheet, "A"+row, "Channel: "+t.ChannelName); err != nil {
					return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
				}
				channelName = t.ChannelName
				i += 2
			}
			row = strconv.Itoa(i)

			if err := f.SetCellValue(sheet, "A"+row, t.TicketType); err != nil {
				return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
			}
			if err := f.SetCellValue(sheet, "B"+row, t.TicketData.Number); err != nil {
				return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
			}
			if err := f.SetCellValue(sheet, "C"+row, t.TicketData.ShortDescription); err != nil {
				return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
			}
			if err := f.SetCellValue(sheet, "D"+row, t.TicketData.Location); err != nil {
				return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
			}
		}

		// Save Excel file
		if err := f.SaveAs(filename); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not save file '%s'", filename)
		}

		g.logger.Infow("Excel file generated", "for", email, "open tickets", len(userTickets))
	}

	return nil
}
