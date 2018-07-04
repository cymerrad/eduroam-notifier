package mailer

import (
	"testing"

	"github.com/google/uuid"
)

func TestSend(t *testing.T) {
	type args struct {
		serverAddr string
		from       string
		to         string
		subject    string
		body       string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test1", args{"localhost:25", "no-reply@localhost", "radek@localhost", "This is a test", "Hello there\n" + uuid.Must(uuid.NewUUID()).String()}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Send(tt.args.serverAddr, tt.args.from, tt.args.to, tt.args.subject, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("SendMail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
