package email

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type Sender interface {
	// SendEmailsForFieldEngineers sends emails with tickets' info to field engineers
	SendEmailsForFieldEngineers(ctx context.Context) error

	// SendEmailsForServiceDesk sends emails with tickets' info to service desk agents
	SendEmailsForServiceDesk(ctx context.Context) error
}

//go:embed email_template.html
var templateHTML string

// NewEmailSender returns new service for sending emails with attached Excel files generated in precious step
func NewEmailSender(
	logger *zap.SugaredLogger,
	postmarkServerURL, postmarkServerToken, messageStream, fromEmailAddress, feAttachmentsDirPath, sdAttachmentsDirPath string,
	ticketRepository repository.TicketRepository, sdAgentEmails []string,
) Sender {
	return &sender{
		logger:               logger,
		postmarkServerURL:    postmarkServerURL,
		postmarkServerToken:  postmarkServerToken,
		messageStream:        messageStream,
		fromEmailAddress:     fromEmailAddress,
		feAttachmentsDirPath: feAttachmentsDirPath,
		sdAttachmentsDirPath: sdAttachmentsDirPath,
		ticketRepository:     ticketRepository,
		sdAgentEmails:        sdAgentEmails,
		client:               http.DefaultClient,
	}
}

type sender struct {
	logger               *zap.SugaredLogger
	postmarkServerURL    string
	postmarkServerToken  string
	messageStream        string
	fromEmailAddress     string
	feAttachmentsDirPath string // directory with Excel files for field engineers
	sdAttachmentsDirPath string // directory with Excel files for field engineers
	ticketRepository     repository.TicketRepository
	sdAgentEmails        []string
	client               *http.Client
}

func (s sender) SendEmailsForFieldEngineers(ctx context.Context) error {
	addresses, err := s.ticketRepository.GetDistinctEmailAddresses(ctx)
	if err != nil {
		return err
	}

	s.logger.Info("Sending emails for Field Engineers")

	caption := "<b>Hi, below are open tickets currently assigned to you.</b>"
	return s.sendEmails(ctx, addresses, caption, "", s.feAttachmentsDirPath)
}

func (s sender) SendEmailsForServiceDesk(ctx context.Context) error {
	s.logger.Info("Sending emails for Service Desk agents")

	caption := "<b>Hi, below are all open tickets.</b>"
	subject := "Open tickets report"
	return s.sendEmails(ctx, s.sdAgentEmails, caption, subject, s.sdAttachmentsDirPath)
}

func (s sender) sendEmails(ctx context.Context, addresses []string, caption, subject, attachmentsDir string) error {
	emails, err := s.prepareEmails(addresses, caption, subject, attachmentsDir)
	if err != nil {
		return err
	}

	s.logger.Infof("Emails to send: %d", len(emails))

	if len(emails) == 0 {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.postmarkServerURL, nil)
	if err != nil {
		return err
	}

	reqPayload, err := json.Marshal(emails)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqPayload))

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Postmark-Server-Token", s.postmarkServerToken)

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = res.Body.Close() }()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		var errorPayload Response
		if err := json.Unmarshal(body, &errorPayload); err != nil {
			return err
		}

		return domain.NewErrorf(domain.ErrorCodeUnknown, "Email service returned error: %+v", errorPayload)
	}

	var respPayload []Response

	err = json.Unmarshal(body, &respPayload)

	for _, r := range respPayload {
		if r.ErrorCode != 0 {
			s.logger.Warnw("Email service returned error for email recipient", "error", r)
		}
	}

	return err
}

func (s sender) prepareEmails(addresses []string, caption, subject, attachmentsDir string) ([]Email, error) {
	var emails []Email

	for _, address := range addresses {
		fileName := address + ".xlsx"

		if subject == "" {
			subject = fmt.Sprintf("Open tickets assigned to %s", address)
		}

		filePath := filepath.Join(attachmentsDir, fileName)

		html, err := s.renderHTML(caption, filePath)
		if err != nil {
			return nil, err
		}

		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return emails, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "Could not open attachment file")
		}

		fileDataEnc := base64.StdEncoding.EncodeToString(fileData)

		text := subject

		e := Email{
			From:     s.fromEmailAddress,
			To:       address,
			Subject:  subject,
			HtmlBody: html,
			TextBody: text,
			Attachments: []Attachment{{
				Name:        fileName,
				Content:     fileDataEnc,
				ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			}},
			MessageStream: s.messageStream,
		}

		emails = append(emails, e)
	}

	return emails, nil
}

func (s sender) renderHTML(caption, excelFile string) (string, error) {
	type HTMLData struct {
		Caption template.HTML
		Table   template.HTML
	}

	tmpl, err := template.New("htmlContent").Parse(templateHTML)
	if err != nil {
		return "", err
	}

	f, err := excelize.OpenFile(excelFile)
	if err != nil {
		return "", err
	}

	defer func() { _ = f.Close() }()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return "", err
	}

	var html string
	totalColsNum := 4

	for i, row := range rows {
		if i == 0 {
			continue // skip the first row, similar info was already added to the email
		}

		var htmlRow string
		emptyColsNum := 0
		if len(row) <= totalColsNum {
			emptyColsNum = totalColsNum - len(row)
		}

		for i, colCell := range row {
			if i == totalColsNum {
				break
			}
			htmlRow += "<td>" + colCell + "</td>"
		}

		for i := 0; i < emptyColsNum; i++ {
			htmlRow += "<td>&nbsp;</td>"
		}

		html += "<tr>" + htmlRow + "</tr>"
	}

	var processedHTML bytes.Buffer
	err = tmpl.Execute(&processedHTML, HTMLData{
		Caption: template.HTML(caption),
		Table:   template.HTML(html),
	})
	if err != nil {
		return "", err
	}

	return processedHTML.String(), nil
}
