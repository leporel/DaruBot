package numbers

import "testing"

func TestNumberRoundTo(t *testing.T) {
	type args struct {
		number int
		to     int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"1", args{1, 5}, 5},
		{"2", args{7, 5}, 10},
		{"3", args{15, 6}, 18},
		{"4", args{5, 5}, 5},
		{"5", args{0, 5}, 5},
		{"6", args{0, -5}, -5},
		{"7", args{13, -5}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NumberRoundTo(tt.args.number, tt.args.to); got != tt.want {
				t.Errorf("NumberRoundTo() = %v, want %v", got, tt.want)
			}
		})
	}
}
