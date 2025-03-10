//go:build windows

package win_eventlog

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"io"
	"reflect"
	"testing"
	"unicode/utf16"
)

func TestDecodeUTF16(t *testing.T) {
	testString := "Test String"
	utf16s := utf16.Encode([]rune(testString))
	var bytesUtf16 bytes.Buffer
	writer := io.Writer(&bytesUtf16)
	lb := len(utf16s)
	for i := 0; i < lb; i++ {
		word := make([]byte, 2)
		binary.LittleEndian.PutUint16(word, utf16s[i])
		_, err := writer.Write(word)
		if err != nil {
			t.Errorf("error preparing UTF-16 test string")
			return
		}
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "Wrong UTF-16",
			args:    args{b: append(bytesUtf16.Bytes(), byte('\x00'))},
			wantErr: true,
		},
		{
			name: "UTF-16",
			args: args{b: bytesUtf16.Bytes()},
			want: []byte(testString),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeUTF16(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeUTF16() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeUTF16() = %v, want %v", got, tt.want)
			}
		})
	}
}

var xmlbroken = `
<BrokenXML>
  <Data/>qq</Data>
</BrokenXML>
`

var xmldata = `
<Event xmlns="http://schemas.microsoft.com/win/2004/08/events/event">
  <UserData>
    <CbsPackageChangeState xmlns="http://manifests.microsoft.com/win/2004/08/windows/setup_provider">
      <IntendedPackageState>5111</IntendedPackageState>
      <ErrorCode><Code>0x0</Code></ErrorCode>
    </CbsPackageChangeState>
  </UserData>
  <EventData>
    <Data>2120-07-26T15:24:25Z</Data>
    <Data>RulesEngine</Data>
    <Data Name="Engine">RulesEngine</Data>
  </EventData>
</Event>
`

type testEvent struct {
	UserData struct {
		InnerXML []byte `xml:",innerxml"`
	} `xml:"UserData"`
	EventData struct {
		InnerXML []byte `xml:",innerxml"`
	} `xml:"EventData"`
}

func TestUnrollXMLFields(t *testing.T) {
	container := testEvent{}
	err := xml.Unmarshal([]byte(xmldata), &container)
	if err != nil {
		t.Errorf("couldn't unmarshal precooked xml string xmldata")
		return
	}

	type args struct {
		data        []byte
		fieldsUsage map[string]int
	}
	tests := []struct {
		name  string
		args  args
		want1 []eventField
		want2 map[string]int
	}{
		{
			name: "Broken XML",
			args: args{
				data:        []byte(xmlbroken),
				fieldsUsage: map[string]int{},
			},
			want1: nil,
			want2: map[string]int{},
		},
		{
			name: "EventData with non-unique names and one Name attr",
			args: args{
				data:        container.EventData.InnerXML,
				fieldsUsage: map[string]int{},
			},
			want1: []eventField{
				{Name: "Data", Value: "2120-07-26T15:24:25Z"},
				{Name: "Data", Value: "RulesEngine"},
				{Name: "Data_Engine", Value: "RulesEngine"},
			},
			want2: map[string]int{"Data": 2, "Data_Engine": 1},
		},
		{
			name: "UserData with non-unique names and three levels of depth",
			args: args{
				data:        container.UserData.InnerXML,
				fieldsUsage: map[string]int{},
			},
			want1: []eventField{
				{Name: "CbsPackageChangeState_IntendedPackageState", Value: "5111"},
				{Name: "CbsPackageChangeState_ErrorCode_Code", Value: "0x0"},
			},
			want2: map[string]int{
				"CbsPackageChangeState_ErrorCode_Code":       1,
				"CbsPackageChangeState_IntendedPackageState": 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := unrollXMLFields(tt.args.data, tt.args.fieldsUsage, "_")
			if !reflect.DeepEqual(got, tt.want1) {
				t.Errorf("ExtractFields() got = %v, want %v", got, tt.want1)
			}
			if !reflect.DeepEqual(got1, tt.want2) {
				t.Errorf("ExtractFields() got1 = %v, want %v", got1, tt.want2)
			}
		})
	}
}

func TestUniqueFieldNames(t *testing.T) {
	type args struct {
		fields      []eventField
		fieldsUsage map[string]int
	}
	tests := []struct {
		name string
		args args
		want []eventField
	}{
		{
			name: "Unique values",
			args: args{
				fields: []eventField{
					{Name: "Data", Value: "2120-07-26T15:24:25Z"},
					{Name: "Data", Value: "RulesEngine"},
					{Name: "Engine", Value: "RulesEngine"},
				},
				fieldsUsage: map[string]int{"Data": 2, "Engine": 1},
			},
			want: []eventField{
				{Name: "Data_1", Value: "2120-07-26T15:24:25Z"},
				{Name: "Data_2", Value: "RulesEngine"},
				{Name: "Engine", Value: "RulesEngine"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := uniqueFieldNames(tt.args.fields, tt.args.fieldsUsage, "_"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PrintFields() = %v, want %v", got, tt.want)
			}
		})
	}
}
