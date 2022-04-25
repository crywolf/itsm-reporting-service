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
	// SendEmails sends emails with tickets' info
	SendEmails(ctx context.Context) error
}

//go:embed email_template.html
var templateHTML string

func NewEmailSender(
	logger *zap.SugaredLogger,
	postmarkServerURL, postmarkServerToken, messageStream, fromEmailAddress, attachmentsDir string,
	ticketRepository repository.TicketRepository,
) Sender {
	return &sender{
		logger:              logger,
		postmarkServerURL:   postmarkServerURL,
		postmarkServerToken: postmarkServerToken,
		messageStream:       messageStream,
		fromEmailAddress:    fromEmailAddress,
		attachmentsDir:      attachmentsDir,
		ticketRepository:    ticketRepository,
		client:              http.DefaultClient,
	}
}

type sender struct {
	logger              *zap.SugaredLogger
	postmarkServerURL   string
	postmarkServerToken string
	messageStream       string
	fromEmailAddress    string
	attachmentsDir      string
	ticketRepository    repository.TicketRepository
	client              *http.Client
}

func (s sender) SendEmails(ctx context.Context) error {
	emails, err := s.prepareEmails(ctx)
	if err != nil {
		return err
	}

	s.logger.Infof("Emails to send: %d", len(emails))

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

func (s sender) prepareEmails(ctx context.Context) ([]Email, error) {
	var emails []Email

	addresses, err := s.ticketRepository.GetDistinctEmails(ctx)
	if err != nil {
		return emails, err
	}

	for _, address := range addresses {
		excelFile := address + ".xlsx"

		subject := fmt.Sprintf("Open tickets assigned to %s", address)
		caption := "<b>Hi, below are open tickets currently assigned to you.</b>"

		html, err := s.renderHTML(caption, excelFile)
		if err != nil {
			return nil, err
		}

		file := filepath.Join(s.attachmentsDir, excelFile)
		fileData, err := os.ReadFile(file)
		if err != nil {
			return emails, err
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
				Name:        excelFile,
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
			continue // skip the first row, similar info was already added to the emailS
		}

		var htmlRow string
		emptyColsNum := 0
		if len(row) <= totalColsNum {
			emptyColsNum = totalColsNum - len(row)
		}

		for _, colCell := range row {
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
