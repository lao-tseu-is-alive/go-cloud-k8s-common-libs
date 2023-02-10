package tools

import (
	"testing"
)

func TestGetOpenApiType(t *testing.T) {
	type args struct {
		t any
	}
	var emptyStringPointer *string
	tests := []struct {
		name    string
		args    args
		wantRes string
		wantErr error
	}{
		{name: "it should return string for a empty string",
			args:    args{t: ""},
			wantRes: "string",
			wantErr: nil,
		},
		{name: "it should return string pointer for a empty string pointer",
			args:    args{t: emptyStringPointer},
			wantRes: "string,nullable",
			wantErr: nil,
		},
		{name: "it should return string for a string with only space",
			args:    args{t: "  "},
			wantRes: "string",
			wantErr: nil,
		},
		{name: "it should return int for a zero",
			args:    args{t: 0},
			wantRes: "integer",
			wantErr: nil,
		},
		{name: "it should return int for a negative number",
			args:    args{t: -10},
			wantRes: "integer",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, gotErr := GetOpenApiType(tt.args.t)
			// fmt.Printf("##GetType(%v)[%T] returns: (%v, error: %v), wants: (%v, error:%v)", tt.args.t, tt.args.t, gotRes, gotErr, tt.wantRes, tt.wantErr)
			if gotRes != tt.wantRes || gotErr != tt.wantErr {
				t.Errorf("GetType(%v)[%T] got: (%v, error: %v), wants: (%v, error:%v)",
					tt.args.t, tt.args.t, gotRes, gotErr, tt.wantRes, tt.wantErr)
			}
		})
	}
}
