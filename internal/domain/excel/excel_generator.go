package excel

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

// Generator is an Excel files generating service
type Generator interface {
	// GenerateExcelFilesForFieldEngineers creates Excel spreadsheet files with tickets' info for field engineers
	GenerateExcelFilesForFieldEngineers(ctx context.Context) error

	// GenerateExcelFilesForServiceDesk creates Excel spreadsheet files with tickets' info for service desk agents
	GenerateExcelFilesForServiceDesk(ctx context.Context) error

	// FEDirPath returns the absolute path to the directory where the files for field engineers are generated to
	FEDirPath() string

	// SDDirPath returns the absolute path to the directory where the files for service desk agents are generated to
	SDDirPath() string
}

// NewExcelGenerator returns new Excel files generating service
func NewExcelGenerator(logger *zap.SugaredLogger, ticketRepository repository.TicketRepository, sdAgentEmails []string) Generator {
	return &excelGen{
		logger:           logger,
		ticketRepository: ticketRepository,
		sdAgentEmails:    sdAgentEmails,
		dirName:          filepath.Join(os.TempDir(), "reporting-xls-files"),
		feSubDir:         "fe",
		sdSubDir:         "sd",
	}
}

type excelGen struct {
	logger           *zap.SugaredLogger
	ticketRepository repository.TicketRepository
	sdAgentEmails    []string
	dirName          string // directory to put generated files to
	feSubDir         string // subdirectory with files for field engineers
	sdSubDir         string // subdirectory with files for service desk agents
}

func (g excelGen) FEDirPath() string {
	return filepath.Join(g.dirName, g.feSubDir)
}

func (g excelGen) SDDirPath() string {
	return filepath.Join(g.dirName, g.sdSubDir)
}

func (g excelGen) GenerateExcelFilesForFieldEngineers(ctx context.Context) error {
	emails, err := g.ticketRepository.GetDistinctEmailAddresses(ctx)
	if err != nil {
		return err
	}

	return g.generateExcelFilesForFE(ctx, emails)
}

func (g excelGen) GenerateExcelFilesForServiceDesk(ctx context.Context) error {
	emails := g.sdAgentEmails
	return g.generateExcelFilesForSD(ctx, emails)
}

func (g excelGen) prepareDirForFE() error {
	return g.prepareDir(g.FEDirPath())
}

func (g excelGen) prepareDirForSD() error {
	return g.prepareDir(g.SDDirPath())
}

func (g excelGen) prepareDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not remove directory '%s'", dir)
	}

	if err := os.MkdirAll(dir, 0750); err != nil {
		return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not create directory '%s'", dir)
	}

	if err := os.Chdir(dir); err != nil {
		return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not change to directory '%s'", dir)
	}

	return nil
}

func (g excelGen) generateExcelFilesForFE(ctx context.Context, emails []string) error {
	if err := g.prepareDirForFE(); err != nil {
		return err
	}

	for _, email := range emails {
		userTickets, err := g.ticketRepository.GetTicketsByEmailAddress(ctx, email)
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
		if err := f.SetColWidth(sheet, "C", "C", 10); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetColWidth(sheet, "D", "E", 45); err != nil {
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
		if err := f.SetCellValue(sheet, "C3", "State"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "D3", "Short description"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "E3", "Location"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}

		if err := g.addRowsWithTicketData(f, userTickets, sheet, filename, false); err != nil {
			return err
		}

		// Save Excel file
		if err := f.SaveAs(filename); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not save file '%s'", filename)
		}

		g.logger.Infow("Excel file for FE generated", "for", email, "open tickets", len(userTickets))
	}

	return nil
}

func (g excelGen) generateExcelFilesForSD(ctx context.Context, emails []string) error {
	if err := g.prepareDirForSD(); err != nil {
		return err
	}

	channelIDs, err := g.ticketRepository.GetDistinctChannelIDs(ctx)
	if err != nil {
		return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not get channel IDs from the ticket repository")
	}

	var channelTickets ticket.List
	for _, channelID := range channelIDs {
		tickets, err := g.ticketRepository.GetTicketsByChannelID(ctx, channelID)
		if err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not get tickets for channel '%s' from the ticket repository", channelID)
		}

		channelTickets = append(channelTickets, tickets...)
	}

	if len(channelTickets) == 0 { // nothing to send
		return nil
	}

	for _, email := range emails {
		filename := email + ".xlsx"

		f := excelize.NewFile()

		// Set columns width
		sheet := "Sheet1"
		if err := f.SetColWidth(sheet, "A", "B", 13); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetColWidth(sheet, "C", "C", 10); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetColWidth(sheet, "D", "E", 45); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetColWidth(sheet, "F", "I", 20); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}

		// Excel file header
		if err := f.SetCellValue(sheet, "A1", "All open tickets"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "A3", "Ticket type"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "B3", "Number"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "C3", "State"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "D3", "Short description"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "E3", "Location"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "F3", "Assigned to (Name)"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "G3", "Assigned to (Email)"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "H3", "Assigned to (Org)"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "I3", "Created at"); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}

		if err := g.addRowsWithTicketData(f, channelTickets, sheet, filename, true); err != nil {
			return err
		}

		// Save Excel file
		if err := f.SaveAs(filename); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not save file '%s'", filename)
		}

		g.logger.Infow("Excel file for SD generated", "for", email, "open tickets", len(channelTickets))
	}

	return nil
}

func (g excelGen) addRowsWithTicketData(f *excelize.File, userTickets ticket.List, sheet string, filename string, withAssignedTo bool) error {
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
		if err := f.SetCellValue(sheet, "C"+row, t.TicketData.StateName()); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "D"+row, t.TicketData.ShortDescription); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}
		if err := f.SetCellValue(sheet, "E"+row, t.TicketData.Location); err != nil {
			return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
		}

		if withAssignedTo {
			if t.UserEmail != "" {
				if err := f.SetCellValue(sheet, "F"+row, t.UserName); err != nil {
					return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
				}
				if err := f.SetCellValue(sheet, "G"+row, t.UserEmail); err != nil {
					return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
				}
				if err := f.SetCellValue(sheet, "H"+row, t.UserOrgName); err != nil {
					return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
				}
			}

			if err := f.SetCellValue(sheet, "I"+row, t.TicketData.CreatedAtDate()); err != nil {
				return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not write data to Excel file '%s'", filename)
			}
		}
	}

	return nil
}
