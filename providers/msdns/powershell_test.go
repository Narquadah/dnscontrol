package msdns

import (
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func Test_generatePSZoneAll(t *testing.T) {
	type args struct {
		dnsserver string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "local",
			args: args{},
			want: `Get-DnsServerZone | ConvertTo-Json`,
		},
		{
			name: "remote",
			args: args{dnsserver: "mydnsserver"},
			want: `Get-DnsServerZone -ComputerName "mydnsserver" | ConvertTo-Json`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generatePSZoneAll(tt.args.dnsserver); got != tt.want {
				t.Errorf("generatePSZoneAll() = got=(\n%s\n) want=(\n%s\n)", got, tt.want)
			}
		})
	}
}

func Test_generatePSZoneDump(t *testing.T) {
	type args struct {
		domainname string
		dnsserver  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "local",
			args: args{domainname: "example.com"},
			want: `$OutputEncoding = [Text.UTF8Encoding]::UTF8 ; Get-DnsServerResourceRecord -ZoneName "example.com" | Select-Object -Property * -ExcludeProperty Cim* | ConvertTo-Json -depth 4 | Out-File "foo" -Encoding utf8`,
		},
		{
			name: "remote",
			args: args{domainname: "example.com", dnsserver: "mydnsserver"},
			want: `$OutputEncoding = [Text.UTF8Encoding]::UTF8 ; Get-DnsServerResourceRecord -ComputerName "mydnsserver" -ZoneName "example.com" | Select-Object -Property * -ExcludeProperty Cim* | ConvertTo-Json -depth 4 | Out-File "foo" -Encoding utf8`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generatePSZoneDump(tt.args.dnsserver, tt.args.domainname, "foo"); got != tt.want {
				t.Errorf("generatePSZoneDump() = got=(\n%s\n) want=(\n%s\n)", got, tt.want)
			}
		})
	}
}

//func Test_generatePSDelete(t *testing.T) {
//	type args struct {
//		domain string
//		rec    *models.RecordConfig
//	}
//	tests := []struct {
//		name string
//		args args
//		want string
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := generatePSDelete(tt.args.domain, tt.args.rec); got != tt.want {
//				t.Errorf("generatePSDelete() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

// func Test_generatePSCreate(t *testing.T) {
// 	type args struct {
// 		domain string
// 		rec    *models.RecordConfig
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := generatePSCreate(tt.args.domain, tt.args.rec); got != tt.want {
// 				t.Errorf("generatePSCreate() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_generatePSModify(t *testing.T) {

	recA1 := &models.RecordConfig{
		Type: "A",
		Name: "@",
	}
	recA1.SetTarget("1.2.3.4")
	recA2 := &models.RecordConfig{
		Type: "A",
		Name: "@",
	}
	recA2.SetTarget("10.20.30.40")

	recMX1 := &models.RecordConfig{
		Type:         "MX",
		Name:         "@",
		MxPreference: 5,
	}
	recMX1.SetTarget("foo.com.")
	recMX2 := &models.RecordConfig{
		Type:         "MX",
		Name:         "@",
		MxPreference: 50,
	}
	recMX2.SetTarget("foo2.com.")

	type args struct {
		domain    string
		dnsserver string
		old       *models.RecordConfig
		rec       *models.RecordConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "A", args: args{domain: "example.com", dnsserver: "", old: recA1, rec: recA2},
			want: `echo DELETE "A" "@" "[target]" ; Remove-DnsServerResourceRecord -Force -ZoneName "example.com" -Name "@" -RRType "A" -RecordData "1.2.3.4" ; echo CREATE "A" "@" "[target]" ; Add-DnsServerResourceRecord -ZoneName "example.com" -Name "@" -TimeToLive $(New-TimeSpan -Seconds 0) -A -IPv4Address "10.20.30.40"`,
		},
		{name: "MX1", args: args{domain: "example.com", dnsserver: "", old: recMX1, rec: recMX2},
			want: `echo DELETE "MX" "@" "[target]" ; Remove-DnsServerResourceRecord -Force -ZoneName "example.com" -Name "@" -RRType "MX" -RecordData 5,"foo.com." ; echo CREATE "MX" "@" "[target]" ; Add-DnsServerResourceRecord -ZoneName "example.com" -Name "@" -TimeToLive $(New-TimeSpan -Seconds 0) -MX -MailExchange "foo2.com." -Preference 50`,
		},
		{name: "A-remote", args: args{domain: "example.com", dnsserver: "myremote", old: recA1, rec: recA2},
			want: `echo DELETE "A" "@" "[target]" ; Remove-DnsServerResourceRecord -ComputerName "myremote" -Force -ZoneName "example.com" -Name "@" -RRType "A" -RecordData "1.2.3.4" ; echo CREATE "A" "@" "[target]" ; Add-DnsServerResourceRecord -ComputerName "myremote" -ZoneName "example.com" -Name "@" -TimeToLive $(New-TimeSpan -Seconds 0) -A -IPv4Address "10.20.30.40"`,
		},
		{name: "MX1-remote", args: args{domain: "example.com", dnsserver: "yourremote", old: recMX1, rec: recMX2},
			want: `echo DELETE "MX" "@" "[target]" ; Remove-DnsServerResourceRecord -ComputerName "yourremote" -Force -ZoneName "example.com" -Name "@" -RRType "MX" -RecordData 5,"foo.com." ; echo CREATE "MX" "@" "[target]" ; Add-DnsServerResourceRecord -ComputerName "yourremote" -ZoneName "example.com" -Name "@" -TimeToLive $(New-TimeSpan -Seconds 0) -MX -MailExchange "foo2.com." -Preference 50`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generatePSModify(tt.args.dnsserver, tt.args.domain, tt.args.old, tt.args.rec); strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("generatePSModify() = got=(\n%s\n) want=(\n%s\n)", got, tt.want)
			}
		})
	}
}
