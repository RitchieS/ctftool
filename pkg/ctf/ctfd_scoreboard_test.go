package ctf

import (
	"reflect"
	"testing"
)

func TestTopTeamData_GetTeam(t *testing.T) {
	type fields struct {
		Num1  Team
		Num2  Team
		Num3  Team
		Num4  Team
		Num5  Team
		Num6  Team
		Num7  Team
		Num8  Team
		Num9  Team
		Num10 Team
	}
	type args struct {
		number int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Team
		wantErr bool
	}{
		{
			name: "get team 1",
			fields: fields{
				Num1: Team{
					Name: "team 1",
				},
			},
			args: args{
				number: 1,
			},
			want: &Team{
				Name: "team 1",
			},
			wantErr: false,
		},
		{
			name: "get team 2",
			fields: fields{
				Num2: Team{
					Name: "team 2",
				},
			},
			args: args{
				number: 2,
			},
			want: &Team{
				Name: "team 2",
			},
			wantErr: false,
		},
		{
			name: "get team 3",
			fields: fields{
				Num3: Team{
					Name: "team 3",
				},
			},
			args: args{
				number: 3,
			},
			want: &Team{
				Name: "team 3",
			},
			wantErr: false,
		},
		{
			name: "get team 4",
			fields: fields{
				Num4: Team{
					Name: "team 4",
				},
			},
			args: args{
				number: 4,
			},
			want: &Team{
				Name: "team 4",
			},
			wantErr: false,
		},
		{
			name: "get team 5",
			fields: fields{
				Num5: Team{
					Name: "team 5",
				},
			},
			args: args{
				number: 5,
			},
			want: &Team{
				Name: "team 5",
			},
			wantErr: false,
		},
		{
			name: "get team 6",
			fields: fields{
				Num6: Team{
					Name: "team 6",
				},
			},
			args: args{
				number: 6,
			},
			want: &Team{
				Name: "team 6",
			},
			wantErr: false,
		},
		{
			name: "get team 7",
			fields: fields{
				Num7: Team{
					Name: "team 7",
				},
			},
			args: args{
				number: 7,
			},
			want: &Team{
				Name: "team 7",
			},
			wantErr: false,
		},
		{
			name: "get team 8",
			fields: fields{
				Num8: Team{
					Name: "team 8",
				},
			},
			args: args{
				number: 8,
			},
			want: &Team{
				Name: "team 8",
			},
			wantErr: false,
		},
		{
			name: "get team 9",
			fields: fields{
				Num9: Team{
					Name: "team 9",
				},
			},
			args: args{
				number: 9,
			},
			want: &Team{
				Name: "team 9",
			},
			wantErr: false,
		},
		{
			name: "get team 10",
			fields: fields{
				Num10: Team{
					Name: "team 10",
				},
			},
			args: args{
				number: 10,
			},
			want: &Team{
				Name: "team 10",
			},
			wantErr: false,
		},
		{
			name: "get team 11",
			fields: fields{
				Num1: Team{
					Name: "team 1",
				},
				Num2: Team{
					Name: "team 2",
				},
				Num3: Team{
					Name: "team 3",
				},
				Num4: Team{
					Name: "team 4",
				},
				Num5: Team{
					Name: "team 5",
				},
				Num6: Team{
					Name: "team 6",
				},
				Num7: Team{
					Name: "team 7",
				},
				Num8: Team{
					Name: "team 8",
				},
				Num9: Team{
					Name: "team 9",
				},
				Num10: Team{
					Name: "team 10",
				},
			},
			args: args{
				number: 11,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &TopTeamData{
				Num1:  tt.fields.Num1,
				Num2:  tt.fields.Num2,
				Num3:  tt.fields.Num3,
				Num4:  tt.fields.Num4,
				Num5:  tt.fields.Num5,
				Num6:  tt.fields.Num6,
				Num7:  tt.fields.Num7,
				Num8:  tt.fields.Num8,
				Num9:  tt.fields.Num9,
				Num10: tt.fields.Num10,
			}
			got, err := d.GetTeam(tt.args.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("TopTeamData.GetTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TopTeamData.GetTeam() = %v, want %v", got, tt.want)
			}
		})
	}
}
