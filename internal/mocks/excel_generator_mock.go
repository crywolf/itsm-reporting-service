package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// ExcelGeneratorMock is an Excel generator mock
type ExcelGeneratorMock struct {
	mock.Mock
}

func (m *ExcelGeneratorMock) GenerateExcelFiles(_ context.Context) error {
	return nil
}

func (m *ExcelGeneratorMock) DirName() string {
	//TODO implement me
	panic("implement me")
}
