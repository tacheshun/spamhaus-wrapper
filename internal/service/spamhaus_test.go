package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

type mockResolver struct {
	lookupIPFunc func(ctx context.Context, network, host string) ([]net.IP, error)
}

func (r *mockResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return r.lookupIPFunc(ctx, network, host)
}

func TestSpamhausService_LookupIP(t *testing.T) {
	tests := []struct {
		name           string
		ip             string
		mockResponses  [][]net.IP
		mockErrors     []error
		expectedResult string
		expectedError  string
	}{
		{
			name: "Valid IP with single response",
			ip:   "127.0.0.2",
			mockResponses: [][]net.IP{
				{net.ParseIP("127.0.0.2")},
			},
			mockErrors:     []error{nil},
			expectedResult: "127.0.0.2",
		},
		{
			name: "Valid IP with multiple responses",
			ip:   "127.0.0.2",
			mockResponses: [][]net.IP{
				{net.ParseIP("127.0.0.2"), net.ParseIP("127.0.0.4"), net.ParseIP("127.0.0.10")},
			},
			mockErrors:     []error{nil},
			expectedResult: "127.0.0.2",
		},
		{
			name:           "Not found IP",
			ip:             "127.0.0.1",
			mockResponses:  [][]net.IP{nil},
			mockErrors:     []error{&net.DNSError{IsNotFound: true}},
			expectedResult: "",
		},
		{
			name:           "Non-determinative response with retry",
			ip:             "127.0.0.3",
			mockResponses:  [][]net.IP{{net.ParseIP("127.255.255.254")}, {net.ParseIP("127.0.0.3")}},
			mockErrors:     []error{nil, nil},
			expectedResult: "127.0.0.3",
		},
		{
			name:          "Persistent non-determinative response",
			ip:            "127.0.0.4",
			mockResponses: [][]net.IP{{net.ParseIP("127.255.255.254")}, {net.ParseIP("127.255.255.254")}, {net.ParseIP("127.255.255.254")}, {net.ParseIP("127.255.255.254")}, {net.ParseIP("127.255.255.254")}},
			mockErrors:    []error{nil, nil, nil, nil, nil},
			expectedError: "received non-determinative response after retries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			mockResolver := &mockResolver{
				lookupIPFunc: func(ctx context.Context, network, host string) ([]net.IP, error) {
					require.Less(t, callCount, len(tt.mockResponses), "Unexpected number of LookupIP calls")
					resp := tt.mockResponses[callCount]
					err := tt.mockErrors[callCount]
					callCount++
					return resp, err
				},
			}

			service := &SpamhausService{resolver: mockResolver}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := service.LookupIP(ctx, tt.ip)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestReverseIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.0.1", "1.0.168.192"},
		{"10.0.0.1", "1.0.0.10"},
		{"172.16.0.1", "1.0.16.172"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := reverseIP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
