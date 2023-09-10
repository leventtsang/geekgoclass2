package service

import (
	"context"
	"errors"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"testing"
)

type MockSMSService struct {
	SendFunc func(ctx context.Context, tplId string, param []string, phones ...string) error
}

func (m *MockSMSService) Send(ctx context.Context, tplId string, param []string, phones ...string) error {
	return m.SendFunc(ctx, tplId, param, phones...)
}

type MockCodeRepository struct {
	StoreFunc  func(ctx context.Context, biz, phone, code string) error
	VerifyFunc func(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

func (m *MockCodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return m.StoreFunc(ctx, biz, phone, code)
}

func (m *MockCodeRepository) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return m.VerifyFunc(ctx, biz, phone, inputCode)
}

func TestSMSCodeService_Send(t *testing.T) {
	tests := []struct {
		name     string
		sendErr  error
		storeErr error
		wantErr  error
	}{
		{
			"Successful send",
			nil,
			nil,
			nil,
		},
		{
			"Error on storing code",
			nil,
			errors.New("store error"),
			errors.New("store error"),
		},
		{
			"Error on sending SMS",
			errors.New("send error"),
			nil,
			errors.New("send error"),
		},
		// More cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSMS := &MockSMSService{
				SendFunc: func(ctx context.Context, tplId string, param []string, phones ...string) error {
					if len(phones) == 0 {
						return errors.New("no phone provided")
					}
					phone := phones[0]
					if phone != "13800000000" {
						return errors.New("unexpected phone number")
					}
					return tt.sendErr
				},
			}

			mockRepo := &MockCodeRepository{
				StoreFunc: func(ctx context.Context, biz, phone, code string) error {
					return tt.storeErr
				},
			}
			service := NewSMSCodeService(mockSMS, mockRepo)

			err := service.Send(context.TODO(), "test", "13800000000")
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("wanted error %v, got %v", tt.wantErr, err)
			}

			err = service.Send(context.TODO(), "test", "13265983381")
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("wanted error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestSMSCodeService_Verify(t *testing.T) {
	tests := []struct {
		name      string
		verifyOk  bool
		verifyErr error
		wantOk    bool
		wantErr   error
	}{
		{
			"Successful verify",
			true,
			nil,
			true,
			nil,
		},
		{
			"Verify error too many times",
			false,
			repository.ErrCodeVerifyTooManyTimes,
			false,
			nil,
		},
		{
			"Other verify error",
			false,
			errors.New("other error"),
			false,
			errors.New("other error"),
		},
		// More cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSMS := &MockSMSService{
				SendFunc: func(ctx context.Context, tplId string, param []string, phones ...string) error {
					if len(phones) == 0 {
						return errors.New("no phone provided")
					}
					phone := phones[0]
					if phone != "13800000000" {
						return errors.New("unexpected phone number")
					}
					return nil
				},
			}

			mockRepo := &MockCodeRepository{
				VerifyFunc: func(ctx context.Context, biz, phone, inputCode string) (bool, error) {
					return tt.verifyOk, tt.verifyErr
				},
			}

			service := NewSMSCodeService(mockSMS, mockRepo)

			ok, err := service.Verify(context.TODO(), "test", "13800000000", "123456")
			if ok != tt.wantOk || (err != nil && err.Error() != tt.wantErr.Error()) {
				t.Errorf("wanted ok %v and error %v, got ok %v and error %v", tt.wantOk, tt.wantErr, ok, err)
			}

			ok, err = service.Verify(context.TODO(), "test", "13265983381", "123456")
			if ok != tt.wantOk || (err != nil && err.Error() != tt.wantErr.Error()) {
				t.Errorf("wanted ok %v and error %v, got ok %v and error %v", tt.wantOk, tt.wantErr, ok, err)
			}
		})
	}
}
