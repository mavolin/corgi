package html

import (
	"testing"
	"time"

	"github.com/mavolin/corgi/v2/internal/should"
)

var D2025_10_01_T_02_03 = time.Date(2025, 10, 1, 2, 3, 0, 0, time.UTC)

func TestMonth(t *testing.T) {
	want := DateTime("2025-10")
	got := Month(D2025_10_01_T_02_03)
	should.Equal(t, want, got)
}

func TestDate(t *testing.T) {
	t.Parallel()

	want := DateTime("2025-10-01")
	got := Date(D2025_10_01_T_02_03)
	should.Equal(t, want, got)
}

func TestYearlessDate(t *testing.T) {
	t.Parallel()

	want := DateTime("10-01")
	got := YearlessDate(D2025_10_01_T_02_03)
	should.Equal(t, want, got)
}

func TestYear(t *testing.T) {
	t.Parallel()

	want := DateTime("2025")
	got := Year(D2025_10_01_T_02_03)
	should.Equal(t, want, got)
}

func TestWeek(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   time.Time
		want DateTime
	}{
		{"mid-year", D2025_10_01_T_02_03, "2025-W40"},
		{
			"iso week year rollover", time.Date(2019, 12, 31, 12, 0, 0, 0, time.UTC), "2020-W01",
		}, // ISO week belongs to 2020
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, Week(c.in), c.want)
		})
	}
}

func TestTime(t *testing.T) {
	t.Parallel()

	T03_04_00_000 := time.Date(0, 1, 1, 3, 4, 0, 0, time.UTC)
	T03_04_05_000 := T03_04_00_000.Add(5 * time.Second)
	T03_04_00_678 := T03_04_00_000.Add(678 * time.Millisecond)
	T03_04_05_678 := T03_04_05_000.Add(678 * time.Millisecond)

	tests := []struct {
		name string
		in   time.Time
		fmt  TimeFormat
		want DateTime
	}{
		{"ToMinute/03:04:00.000", T03_04_00_000, ToMinute, "03:04"},
		{"ToMinute/03:04:05.000", T03_04_05_000, ToMinute, "03:04"},
		{"ToSecond/03:04:00.000", T03_04_00_000, ToSecond, "03:04"},
		{"ToSecond/03:04:05.000", T03_04_05_000, ToSecond, "03:04:05"},
		{"ToMillisecond/03:04:00.000", T03_04_00_000, ToMillisecond, "03:04"},
		{"ToMillisecond/03:04:05.000", T03_04_05_000, ToMillisecond, "03:04:05"},
		{"ToMillisecond/03:04:00.678", T03_04_00_678, ToMillisecond, "03:04:00.678"},
		{"ToMillisecond/03:04:05.678", T03_04_05_678, ToMillisecond, "03:04:05.678"},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, Time(c.in, c.fmt), c.want)
		})
	}
}

func TestLocalDateTime(t *testing.T) {
	t.Parallel()

	in := time.Date(2025, 10, 22, 15, 4, 3, 0, time.UTC)
	should.Equal(t, LocalDateTime(in, ToSecond), DateTime("2025-10-22T15:04:03"))
}

func TestGlobalDateTime(t *testing.T) {
	t.Parallel()

	date := time.Date(2025, 10, 01, 0, 0, 0, 0, time.UTC)
	wantDate := "2025-10-01"

	times := []struct {
		toMinute, toSecond, toMillisecond string
		add                               time.Duration
	}{
		{"03:04", "03:04", "03:04", 3*time.Hour + 4*time.Minute},
		{"03:04", "03:04:05", "03:04:05", 3*time.Hour + 4*time.Minute + 5*time.Second},
		{"03:04", "03:04", "03:04:00.678", 3*time.Hour + 4*time.Minute + 678*time.Millisecond},
		{"03:04", "03:04:05", "03:04:05.678", 3*time.Hour + 4*time.Minute + 5*time.Second + 678*time.Millisecond},
	}

	for _, tm := range times {
		in := date.Add(tm.add)
		t.Run(tm.toMillisecond, func(t *testing.T) {
			t.Parallel()

			zones := []struct {
				want string
				loc  *time.Location
			}{
				{"Z", time.UTC},
				{"+02", time.FixedZone("+02", 2*60*60)},
				{"+05:30", time.FixedZone("+05:30", (5*60+30)*60)},
				{"-03", time.FixedZone("-03", -3*60*60)},
				{"-03:30", time.FixedZone("-03:30", -(3*60+30)*60)},
				{"InvalidTimeZoneOffset", time.FixedZone("+02:03:30", 2*60*60+3*60+30)},
			}
			for _, zn := range zones {
				t.Run(zn.want, func(t *testing.T) {
					t.Parallel()
					in := time.Date(in.Year(), in.Month(), in.Day(), in.Hour(), in.Minute(), in.Second(), in.Nanosecond(), zn.loc)

					timeFormats := []struct {
						name   string
						format TimeFormat
						want   string
					}{
						{"ToMinute", ToMinute, tm.toMinute},
						{"ToSecond", ToSecond, tm.toSecond},
						{"ToMillisecond", ToMillisecond, tm.toMillisecond},
					}
					for _, tf := range timeFormats {
						t.Run(tf.name, func(t *testing.T) {
							t.Parallel()

							want := DateTime(wantDate + "T" + tf.want + zn.want)
							if zn.want == "InvalidTimeZoneOffset" {
								want = "InvalidTimeZoneOffset"
							}

							got := GlobalDateTime(in, tf.format)
							should.Equal(t, got, want)
						})
					}
				})
			}
		})
	}
}

func TestTimeZoneOffset(t *testing.T) {
	tests := []struct {
		want string
		loc  *time.Location
	}{
		{"Z", time.UTC},
		{"+02", time.FixedZone("+02", 2*60*60)},
		{"+05:30", time.FixedZone("+05:30", (5*60+30)*60)},
		{"-03", time.FixedZone("-03", -3*60*60)},
		{"-03:30", time.FixedZone("-03:30", -(3*60+30)*60)},
		{"InvalidTimeZoneOffset", time.FixedZone("+02:03:30", 2*60*60+3*60+30)},
	}

	for _, c := range tests {
		t.Run(c.want, func(t *testing.T) {
			in := D2025_10_01_T_02_03.In(c.loc)
			should.Equal(t, TimeZoneOffset(in), DateTime(c.want))
		})
	}
}

func TestDuration(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want DateTime
	}{
		{-time.Second, "NegativeDuration"},
		{0, "PT0S"},
		{48 * time.Hour, "P2D"},
		{24*time.Hour + 2*time.Hour, "P1DT2H"},
		{2 * time.Hour, "PT2H"},
		{2*time.Hour + 3*time.Second, "PT2H3S"},
		{3 * time.Minute, "PT3M"},
		{45 * time.Second, "PT45S"},
		{1 * time.Millisecond, "PT0.1S"},
		{12*time.Second + 340*time.Millisecond, "PT12.34S"},
		{1*time.Hour + 2*time.Minute + 3*time.Second + 4*time.Millisecond, "PT1H2M3.4S"},
	}
	for _, c := range tests {
		t.Run(c.in.String(), func(t *testing.T) {
			should.Equal(t, Duration(c.in), c.want)
		})
	}
}
