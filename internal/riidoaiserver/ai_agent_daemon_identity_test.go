package riidoaiserver

import "testing"

func TestSameDaemonIdentityRequiresDeviceAndProfileOrDaemonMatch(t *testing.T) {
	cases := []struct {
		name string
		a    DeviceDaemonRecord
		b    DeviceDaemonRecord
		want bool
	}{
		{
			name: "same profile wins over daemon id",
			a:    DeviceDaemonRecord{DeviceID: " device-1 ", Profile: " staging ", DaemonID: "old"},
			b:    DeviceDaemonRecord{DeviceID: "device-1", Profile: "staging", DaemonID: "new"},
			want: true,
		},
		{
			name: "profile mismatch",
			a:    DeviceDaemonRecord{DeviceID: "device-1", Profile: "staging", DaemonID: "same"},
			b:    DeviceDaemonRecord{DeviceID: "device-1", Profile: "production", DaemonID: "same"},
		},
		{
			name: "same daemon id without profile",
			a:    DeviceDaemonRecord{DeviceID: "device-1", DaemonID: "daemon-1"},
			b:    DeviceDaemonRecord{DeviceID: "device-1", DaemonID: " daemon-1 "},
			want: true,
		},
		{
			name: "device mismatch",
			a:    DeviceDaemonRecord{DeviceID: "device-1", Profile: "staging"},
			b:    DeviceDaemonRecord{DeviceID: "device-2", Profile: "staging"},
		},
		{
			name: "missing identity scope",
			a:    DeviceDaemonRecord{DeviceID: "device-1"},
			b:    DeviceDaemonRecord{DeviceID: "device-1"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sameDaemonIdentity(tc.a, tc.b); got != tc.want {
				t.Fatalf("sameDaemonIdentity() = %v, want %v", got, tc.want)
			}
		})
	}
}
