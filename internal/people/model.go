// Package people innehåller domänmodellen och parsern för det radbaserade formatet.
package people

import "encoding/xml"

// People är rotelementet <people>.
type People struct {
	XMLName xml.Name  `xml:"people"`
	Persons []*Person `xml:"person"`
}

// Person motsvarar en P-rad.
type Person struct {
	FirstName string    `xml:"firstname"`
	LastName  string    `xml:"lastname"`
	Address   *Address  `xml:"address,omitempty"`
	Phone     *Phone    `xml:"phone,omitempty"`
	Family    []*Family `xml:"family"`
}

// Phone motsvarar en T-rad.
type Phone struct {
	Mobile   string `xml:"mobile,omitempty"`
	Landline string `xml:"landline,omitempty"`
}

// Address motsvarar en A-rad.
type Address struct {
	Street string `xml:"street"`
	City   string `xml:"city"`
	Zip    string `xml:"zip,omitempty"`
}

// Family motsvarar en F-rad.
type Family struct {
	Name    string   `xml:"name"`
	Born    string   `xml:"born"`
	Address *Address `xml:"address,omitempty"`
	Phone   *Phone   `xml:"phone,omitempty"`
}

// contactTarget är den gemensamma mottagaren för T- och A-rader.
type contactTarget interface {
	setAddress(*Address)
	setPhone(*Phone)
	hasAddress() bool
	hasPhone() bool
}

// Person och Family implementerar contactTarget nedan.
func (p *Person) setAddress(a *Address) { p.Address = a }
func (p *Person) setPhone(t *Phone)     { p.Phone = t }
func (p *Person) hasAddress() bool      { return p.Address != nil }
func (p *Person) hasPhone() bool        { return p.Phone != nil }

func (f *Family) setAddress(a *Address) { f.Address = a }
func (f *Family) setPhone(t *Phone)     { f.Phone = t }
func (f *Family) hasAddress() bool      { return f.Address != nil }
func (f *Family) hasPhone() bool        { return f.Phone != nil }