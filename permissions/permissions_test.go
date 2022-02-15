package permissions_test

import (
	"testing"

	"awesome-dragon.science/go/irc/permissions"
)

func TestAnyPermissionMatch(t *testing.T) { //nolint:funlen,dupl // tests
	t.Parallel()

	type args struct {
		available []string
		required  []string
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "both nil",
			args: args{
				available: nil,
				required:  nil,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "both empty",
			args: args{
				available: []string{},
				required:  []string{},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "bad glob",
			args: args{
				available: []string{"[]", "b", "c", "d"},
				required:  []string{"a"},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "multiple matches",
			args: args{
				available: []string{"this", "is", "a.*"},
				required:  []string{"this", "is", "a.test"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "off by one",
			args: args{
				available: []string{"if you immediately know the candlelight is fire"},
				required:  []string{"if you immediately know the candlelight is fire", "then the meal was cooked a long time ago"},
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := permissions.AnyPermissionMatch(tt.args.available, tt.args.required)
			if (err != nil) != tt.wantErr {
				t.Errorf("AnyPermissionMatch() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("AnyPermissionMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllPermissionMatch(t *testing.T) { //nolint:funlen,dupl // its a test
	t.Parallel()

	type args struct {
		available []string
		required  []string
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "both nil",
			args: args{
				available: nil,
				required:  nil,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "both empty",
			args: args{
				available: []string{},
				required:  []string{},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "bad glob",
			args: args{
				available: []string{"[]", "b", "c", "d"},
				required:  []string{"a"},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "multiple matches",
			args: args{
				available: []string{"this", "is", "a.*"},
				required:  []string{"this", "is", "a.test"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "off by one",
			args: args{
				available: []string{"if you immediately know the candlelight is fire"},
				required:  []string{"if you immediately know the candlelight is fire", "then the meal was cooked a long time ago"},
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := permissions.AllPermissionMatch(tt.args.available, tt.args.required)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllPermissionMatch() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("AllPermissionMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
