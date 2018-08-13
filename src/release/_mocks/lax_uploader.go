package mocks

import io "io"
import mock "github.com/stretchr/testify/mock"

type LaxUploader struct {
	mock.Mock
}

func (_m *LaxUploader) File(string, io.ReadSeeker) (string, error) {
	return "", nil
}

func (_m *LaxUploader) JSON(string, interface{}) error {
	return nil
}
