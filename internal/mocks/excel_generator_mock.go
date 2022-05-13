package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// ExcelGeneratorMock is an Excel generator mock
type ExcelGeneratorMock struct {
	mock.Mock
}

func (m *ExcelGeneratorMock) GenerateExcelFilesForFieldEngineers(_ context.Context) error {
	args := m.Called()
	return args.Error(0)
}

func (m *ExcelGeneratorMock) GenerateExcelFilesForServiceDesk(_ context.Context) error {
	args := m.Called()
	return args.Error(0)
}

func (m *ExcelGeneratorMock) FEDirPath() string {
	//TODO implement me
	panic("implement me")
}

func (m *ExcelGeneratorMock) SDDirPath() string {
	//TODO implement me
	panic("implement me")
}
