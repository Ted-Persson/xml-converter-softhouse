package people

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const separator = "|"

// ParseError beskriver ett fel på en specifik rad i indata.
type ParseError struct {
	Line int
	Msg  string
}

// Error implementerar Go:s error-interface.
func (e *ParseError) Error() string {
	return fmt.Sprintf("rad %d: %s", e.Line, e.Msg)
}

// parser håller tillståndet som behövs mellan raderna.
type parser struct {
	people  *People
	person  *Person       // senast påbörjade P
	current contactTarget
	line    int
}

// Parse läser det radbaserade formatet från r och bygger upp modellen.
func Parse(r io.Reader) (*People, error) {
	p := &parser{people: &People{}}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		p.line++
		if err := p.handle(scanner.Text()); err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("kunde inte läsa indata: %w", err)
	}
	return p.people, nil
}

// handle dirigerar en rad till rätt funktion baserat på dess första bokstav.
func (p *parser) handle(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil // tomma rader ignoreras
	}

	fields := strings.Split(raw, separator)
	switch fields[0] {
	case "P":
		return p.startPerson(fields)
	case "F":
		return p.startFamily(fields)
	case "A":
		return p.addAddress(fields)
	case "T":
		return p.addPhone(fields)
	default:
		return p.errf("okänd posttyp %q", fields[0])
	}
}

// startPerson öppnar en ny person och stänger allt tidigare öppet.
func (p *parser) startPerson(f []string) error {
	if err := p.requireFields(f, 3, 3, "P|förnamn|efternamn"); err != nil {
		return err
	}
	person := &Person{FirstName: f[1], LastName: f[2]}
	p.people.Persons = append(p.people.Persons, person)
	p.person = person
	p.current = person
	return nil
}

// startFamily lägger en familjemedlem på aktuell person.
func (p *parser) startFamily(f []string) error {
	if err := p.requireFields(f, 3, 3, "F|namn|födelseår"); err != nil {
		return err
	}
	if p.person == nil {
		return p.errf("F-rad utan föregående P-rad")
	}
	family := &Family{Name: f[1], Born: f[2]}
	p.person.Family = append(p.person.Family, family)
	p.current = family
	return nil
}

// addAddress sätter adress på aktuellt mål (person eller familjemedlem).
func (p *parser) addAddress(f []string) error {
	if err := p.requireFields(f, 3, 4, "A|gata|stad[|postnummer]"); err != nil {
		return err
	}
	if p.current == nil {
		return p.errf("A-rad utan föregående P- eller F-rad")
	}
	if p.current.hasAddress() {
		return p.errf("dubbel A-rad för samma post")
	}
	addr := &Address{Street: f[1], City: f[2]}
	if len(f) == 4 {
		addr.Zip = f[3]
	}
	p.current.setAddress(addr)
	return nil
}

// addPhone sätter telefon på aktuellt mål (person eller familjemedlem).
func (p *parser) addPhone(f []string) error {
	if err := p.requireFields(f, 3, 3, "T|mobilnummer|fastnätsnummer"); err != nil {
		return err
	}
	if p.current == nil {
		return p.errf("T-rad utan föregående P- eller F-rad")
	}
	if p.current.hasPhone() {
		return p.errf("dubbel T-rad för samma post")
	}
	p.current.setPhone(&Phone{Mobile: f[1], Landline: f[2]})
	return nil
}

// requireFields kontrollerar att raden har rätt antal fält.
func (p *parser) requireFields(f []string, min, max int, want string) error {
	if len(f) < min || len(f) > max {
		return p.errf("förväntade formatet %s, fick %d fält", want, len(f))
	}
	return nil
}

// errf skapar ett *ParseError med aktuellt radnummer.
func (p *parser) errf(format string, args ...any) error {
	return &ParseError{Line: p.line, Msg: fmt.Sprintf(format, args...)}
}