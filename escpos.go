// escpos project escpos.go
package escpos

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	// "github.com/alexbrainman/printer"
	"github.com/tarm/serial"
)

type escpos struct {
	name       string
	dst        net.Conn
	dst_serial *serial.Port
	dst_usb    *os.File
	//dst_usb_win  *printer.Printer
	under        int
	bold         int
	bw           int
	type_printer int
}

const (
	UPC_A   = 65
	UPC_E   = 66
	EAN13   = 67
	EAN8    = 68
	CODE39  = 69
	I25     = 70
	CODEBAR = 71
	CODE93  = 72
	CODE128 = 73
	CODE11  = 74
	MSI     = 75
)

const (
	ETH   = 1
	SER   = 2
	USB   = 3
	USB_W = 4
)

func NewPrinter() escpos {
	return escpos{}
}

func (e escpos) SetSerialDst(port string, baudrate int, timeout int) error {
	c := &serial.Config{Name: port, Baud: baudrate, ReadTimeout: time.Second * time.Duration(timeout)}
	s, err := serial.OpenPort(c)

	if err != nil {
		return err
	} else {
		e.dst_serial = s
		e.type_printer = SER
		return nil
	}

}

func (e escpos) SetTCPDst(ip, port string, timeout int) error {
	wait := time.Duration(timeout) * time.Second

	cn, err := net.DialTimeout("tcp", ip+":"+port, wait)

	if err != nil {
		return err
	} else {
		e.dst = cn
		e.type_printer = ETH
		return nil
	}
}

func (e escpos) SetUSBDst(printer_path string) error {
	f, err := os.OpenFile(printer_path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	defer f.Close()

	if err != nil {
		return err
	} else {
		e.dst_usb = f
		e.type_printer = USB
		return nil
	}
}

/*
func (e escpos) SetWINDst(printer_name, doc_name string) error {
	p, err := printer.Open(printer_name)
	if err != nil {
		return err
	}
	defer p.Close()

	err = p.StartDocument(doc_name, "RAW")
	if err != nil {
		return err
	}
	defer p.EndDocument()
	err = p.StartPage()
	if err != nil {
		return err
	} else {
		e.dst_usb_win = p
		e.type_printer = USB_W
		return nil
	}

}*/

/**/

func (e escpos) CloseDst() error {
	switch e.type_printer {
	case ETH:
		return e.dst.Close()
	case SER:
		return e.dst_serial.Close()
	case USB:
		return nil
	/*case USB_W:
	//return e.dst_usb_win.EndPage()*/
	default:
		return e.dst.Close()
	}

}

func (e escpos) Send(msg string) {
	switch e.type_printer {
	case ETH:
		fmt.Fprintf(e.dst, msg)
	case SER:
		e.dst_serial.Write([]byte(msg))
	case USB:
		e.dst_usb.Write([]byte(msg))
	case USB_W:
		//fmt.Fprintf(e.dst_usb_win, msg)
	default:
		fmt.Fprintf(e.dst, msg)
	}

}

func (e *escpos) toggleBW() {
	if e.bw == 1 {
		e.bw = 0
	} else {
		e.bw = 1
	}
	t := fmt.Sprintf("\x1DB%c", e.bw)
	e.Send(t)
}

func (e escpos) init() {
	e.Send("\x1B\x40")
}

func (e escpos) cut() {
	e.Send("\x1DVA0")
}

func (e escpos) ff() {
	e.Send("\n")
}
func (e escpos) ffn(n int) {
	t := fmt.Sprintf("\x1Bd%c", n)
	e.Send(t)
}
func (e escpos) codePage858() {
	t := fmt.Sprintf("\x1BR%c", 2)
	e.Send(t)
}
func (e escpos) fontA() {
	e.Send("\x1BM0")
}
func (e escpos) fontB() {
	e.Send("\x1BM1")
}
func (e escpos) fontC() {
	e.Send("\x1B!1")
}

func (e escpos) doubleStrike() {
	e.Send("\x1BG\x01")
}
func (e *escpos) toggleBold() {
	if e.bold == 1 {
		e.bold = 0
	} else {
		e.bold = 1
	}
	t := fmt.Sprintf("\x1BG%c", e.bold)
	e.Send(t)
}

func (e escpos) underline() {
	e.Send("\x1B-\x01")
	e.under = 1
}

func (e *escpos) toggleUnderline() {
	if e.under == 1 {
		e.under = 0
	} else {
		e.under = 1
	}
	t := fmt.Sprintf("\x1B-%c", e.under)
	e.Send(t)
}

func (e escpos) left() {
	e.Send("\x1Ba\x00")
}

func (e escpos) centre() {
	e.Send("\x1Ba\x01")
}

func (e escpos) right() {
	e.Send("\x1Ba\x02")
}

func (e escpos) reallywide() {
	e.Send("\x1D\x21\x70")
}

func (e escpos) normalwide() {
	e.Send("\x1D\x21\x00")
}

func (e escpos) printBarCode(bcs int, hr int, msg string) error {

	var err error
	length := len(msg)
	switch bcs {
	case UPC_A, UPC_E:
		if length < 11 || length > 12 {
			err = errors.New("UPC : Message length must be between 11 and 12 characters")
		}
	case EAN13:
		if length < 12 || length > 13 {
			err = errors.New("EAN13: Message length must be between 12 and 13 characters")
		}
	case EAN8:
		if length < 7 || length > 8 {
			err = errors.New("EAN8: Message length must be between 7 and 8 characters")
		}
	case I25:
		if length < 1 || (length%2 != 0) {
			err = errors.New("I25: Message length must be greater than 1 and even")
		}
	default:
		if length < 1 {
			err = errors.New("Message Length must be greater than 1")
		}

	}

	if err == nil {

		t := ""

		if hr > 0 {
			t = fmt.Sprintf("\x1DH%c", hr)
		}

		t += fmt.Sprintf("\x1Dk%c%c%s", bcs, length, msg)
		e.Send(t)

	}

	return err

}

func (e escpos) printQrCode(model int, size int, ec int, message string) {
	//GS (k pL pH cn fn parameters
	/*
		GS (k 4 0 "1" 65 50 0
		GS (k 3 0 "1" 67 5
		GS (k 3 0 "1" 69 48
		GS (k len(msg)+3 0 "1" 80 48 "msg"
		GS (k 3 0 "1" 81 48
	*/

	//size = 1,2,3,4,5,10,16
	//model = 1,2,3( aka micro)
	//ec = L(1),M(2),Q(3),H(4)

	t := fmt.Sprintf("\x1D\x28k%c%c%s%c%c%c", 4, 0, "1", 65, 48+model, 0)
	t += fmt.Sprintf("\x1D\x28k%c%c%s%c%c", 3, 0, "1", 67, size)
	t += fmt.Sprintf("\x1D\x28k%c%c%s%c%c", 3, 0, "1", 69, 48+ec)
	t += fmt.Sprintf("\x1D\x28k%c%c%s%c%c%s", len(message)+3, 0, "1", 80, 48, message)
	t += fmt.Sprintf("\x1D\x28k%c%c%s%c%c", 3, 0, "1", 81, 48)

	e.Send(t)
}
