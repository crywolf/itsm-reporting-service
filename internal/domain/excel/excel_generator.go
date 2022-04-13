package excel

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
	"github.com/xuri/excelize/v2"
)

type Generator interface {
	// GenerateExcelFiles creates Excel spreadsheet files with tickets info
	GenerateExcelFiles(ctx context.Context) error

	// DirName returns name of the directory where the files are generated to
	DirName() string
}

func NewExcelGenerator(ticketRepository repository.TicketRepository) Generator {
	return &excelGen{
		ticketRepository: ticketRepository,
		dirName:          filepath.Join(os.TempDir(), "xls-files"),
	}
}

type excelGen struct {
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

	// TODO remove all fmt.Prints

	fmt.Println("\n===> GenerateExcelFiles - emails", emails)

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

		fmt.Println("\n\n===> GenerateExcelFiles - email, len(tickets)", email, len(userTickets))
		fmt.Println("\n===> GenerateExcelFiles - tickets", userTickets)

		filename := email + ".xlsx"

		f := excelize.NewFile()

		// Excel file header
		if err := f.SetCellValue("Sheet1", "A1", "Tickets for "+email); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s", filename)
		}

		// Excel rows with ticket data
		for i, t := range userTickets {
			row := strconv.Itoa(i + 2)

			if err := f.SetCellValue("Sheet1", "A"+row, t.TicketType); err != nil {
				return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s", filename)
			}
			if err := f.SetCellValue("Sheet1", "B"+row, t.ChannelName); err != nil {
				return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s", filename)
			}
			if err := f.SetCellValue("Sheet1", "C"+row, t.TicketData.Number); err != nil {
				return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s", filename)
			}
			if err := f.SetCellValue("Sheet1", "D"+row, t.TicketData.ShortDescription); err != nil {
				return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s", filename)
			}
		}

		// Save Excel file
		if err := f.SaveAs(filename); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not save file '%s'", filename)
		}

		fmt.Println("====>", filename, "SAVED!!!")
	}

	return nil
}
