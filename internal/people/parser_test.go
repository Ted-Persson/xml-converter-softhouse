package people

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestParseExample kontrollerar hela kedjan mot exemplet i uppgiften.
// testdata/expected.xml är en "golden file": ändras utdataformatet faller
// testet, vilket är precis vad man vill i en integration mot ett annat system.
func TestParseExample(t *testing.T) {
	in, err := os.Open("../../testdata/example.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()

	model, err := Parse(in)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	var got bytes.Buffer
	if err := WriteXML(&got, model); err != nil {
		t.Fatalf("WriteXML: %v", err)
	}

	want, err := os.ReadFile("../../testdata/expected.xml")
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != string(want) {
		t.Errorf("utdata skiljer sig från expected.xml\n--- fick ---\n%s\n--- ville ha ---\n%s", got.String(), want)
	}
}

// TestPhoneAndAddressBindToLatestContext är kärnan i formatet: T och A hör
// till den senaste F-raden om en sådan finns, annars till P-raden.
func TestPhoneAndAddressBindToLatestContext(t *testing.T) {
	model := mustParse(t, "P|Anna|Andersson\nT|070-1|08-1\nF|Kalle|1990\nT|070-2|08-2\n")

	person := model.Persons[0]
	if person.Phone == nil || person.Phone.Mobile != "070-1" {
		t.Errorf("personens telefon blev %+v", person.Phone)
	}
	if len(person.Family) != 1 {
		t.Fatalf("förväntade 1 familjemedlem, fick %d", len(person.Family))
	}
	if person.Family[0].Phone == nil || person.Family[0].Phone.Mobile != "070-2" {
		t.Errorf("familjemedlemmens telefon blev %+v", person.Family[0].Phone)
	}
}

// TestNewPersonClosesFamilyContext säkerställer att en F-rad inte "läcker"
// vidare till nästa person.
func TestNewPersonClosesFamilyContext(t *testing.T) {
	model := mustParse(t, "P|Anna|Andersson\nF|Kalle|1990\nP|Bo|Bosson\nA|Storgatan 1|Ronneby\n")

	if len(model.Persons) != 2 {
		t.Fatalf("förväntade 2 personer, fick %d", len(model.Persons))
	}
	if len(model.Persons[1].Family) != 0 {
		t.Errorf("andra personen fick oväntade familjemedlemmar")
	}
	if model.Persons[1].Address == nil {
		t.Errorf("adressen hamnade inte på den nya personen")
	}
}

// TestAddressWithoutZip motsvarar Boris-fallet i uppgiften.
func TestAddressWithoutZip(t *testing.T) {
	model := mustParse(t, "P|Boris|Johnson\nA|10 Downing Street|London\n")

	addr := model.Persons[0].Address
	if addr == nil || addr.Zip != "" {
		t.Fatalf("förväntade adress utan postnummer, fick %+v", addr)
	}

	var buf bytes.Buffer
	if err := WriteXML(&buf, model); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "<zip>") {
		t.Errorf("tomt postnummer ska utelämnas helt, inte skrivas som tom tagg")
	}
}

// TestSpecialCharactersAreEscaped visar varför XML:en byggs med
// encoding/xml i stället för strängkonkatenering.
func TestSpecialCharactersAreEscaped(t *testing.T) {
	model := mustParse(t, "P|Tom & Jerry|<Sundin>\n")

	var buf bytes.Buffer
	if err := WriteXML(&buf, model); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Tom &amp; Jerry") || !strings.Contains(out, "&lt;Sundin&gt;") {
		t.Errorf("specialtecken escapades inte: %s", out)
	}
}

// TestEmptyInputGivesEmptyRoot — tom fil är inte ett fel.
func TestEmptyInputGivesEmptyRoot(t *testing.T) {
	model := mustParse(t, "")
	if len(model.Persons) != 0 {
		t.Errorf("förväntade inga personer, fick %d", len(model.Persons))
	}
}

func TestCRLFLineEndings(t *testing.T) {
	model := mustParse(t, "P|Anna|Andersson\r\nA|Storgatan 1|Ronneby|37200\r\n")
	if got := model.Persons[0].Address.Zip; got != "37200" {
		t.Errorf("postnummer blev %q, vagnretur hängde troligen kvar", got)
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantLine int
	}{
		{"okänd posttyp", "P|Anna|Andersson\nX|något\n", 2},
		{"T utan föregående P", "T|070-1|08-1\n", 1},
		{"A utan föregående P", "A|Storgatan 1|Ronneby\n", 1},
		{"F utan föregående P", "F|Kalle|1990\n", 1},
		{"för få fält på P", "P|Anna\n", 1},
		{"för många fält på A", "A|a|b|c|d\n", 1},
		{"dubbel adress", "P|Anna|A\nA|a|b\nA|c|d\n", 3},
		{"dubbel telefon", "P|Anna|A\nT|1|2\nT|3|4\n", 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tc.input))
			if err == nil {
				t.Fatal("förväntade ett fel")
			}
			perr, ok := err.(*ParseError)
			if !ok {
				t.Fatalf("förväntade *ParseError, fick %T", err)
			}
			if perr.Line != tc.wantLine {
				t.Errorf("felet pekade på rad %d, förväntade %d", perr.Line, tc.wantLine)
			}
		})
	}
}

func mustParse(t *testing.T, input string) *People {
	t.Helper()
	model, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return model
}