package services

import (
	"errors"
	"net/smtp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// mockSMTPClient captures the parameters passed to SendMail
type mockSMTPClient struct {
	called    bool
	addr      string
	from      string
	to        []string
	message   []byte
	returnErr error
}

func (m *mockSMTPClient) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	m.called = true
	m.addr = addr
	m.from = from
	m.to = append([]string{}, to...)
	m.message = append([]byte{}, msg...)
	return m.returnErr
}

func TestSendDonationReceipt_UsesClientAndRespectsEmailEnabled(t *testing.T) {
	mock := &mockSMTPClient{}
	es := &EmailService{
		SMTPHost:     "smtp.test",
		SMTPPort:     "1025",
		SMTPUsername: "user",
		SMTPPassword: "pass",
		FromEmail:    "from@test.local",
		FromName:     "Test",
		EmailEnabled: true,
		client:       mock,
	}

	testData := DonationReceiptData{
		DonorName:        "Mock Donor",
		DonationAmount:   10.0,
		DonationType:     "One-time",
		TransactionID:    "MOCK-1",
		DonationDate:     time.Now(),
		OrganizationName: "Test Org",
	}

	err := es.SendDonationReceipt("recipient@test.local", testData)
	require.NoError(t, err)
	require.True(t, mock.called)
	require.Contains(t, mock.addr, "smtp.test")
	require.Contains(t, string(mock.message), "MOCK-1")
}

func TestSendDonationReceipt_WhenDisabled_DoesNotCallClient(t *testing.T) {
	mock := &mockSMTPClient{}
	es := &EmailService{
		SMTPHost:     "smtp.test",
		SMTPPort:     "1025",
		SMTPUsername: "user",
		SMTPPassword: "pass",
		FromEmail:    "from@test.local",
		FromName:     "Test",
		EmailEnabled: false,
		client:       mock,
	}

	testData := DonationReceiptData{
		DonorName:        "Mock Donor",
		DonationAmount:   10.0,
		DonationType:     "One-time",
		TransactionID:    "MOCK-2",
		DonationDate:     time.Now(),
		OrganizationName: "Test Org",
	}

	err := es.SendDonationReceipt("recipient@test.local", testData)
	require.NoError(t, err)
	require.False(t, mock.called, "SMTP client should not be called when EmailEnabled is false")
}

func TestSendDonationReceipt_ClientErrorPropagates(t *testing.T) {
	mock := &mockSMTPClient{returnErr: errors.New("smtp failure")}
	es := &EmailService{
		SMTPHost:     "smtp.test",
		SMTPPort:     "1025",
		SMTPUsername: "user",
		SMTPPassword: "pass",
		FromEmail:    "from@test.local",
		FromName:     "Test",
		EmailEnabled: true,
		client:       mock,
	}

	testData := DonationReceiptData{
		DonorName:        "Mock Donor",
		DonationAmount:   10.0,
		DonationType:     "One-time",
		TransactionID:    "MOCK-3",
		DonationDate:     time.Now(),
		OrganizationName: "Test Org",
	}

	err := es.SendDonationReceipt("recipient@test.local", testData)
	require.Error(t, err)
	require.True(t, mock.called)
}
