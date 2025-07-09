package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// RESP (Redis Serialization Protocol) constants
const (
	SimpleString = '+'
	Error        = '-'
	Integer      = ':'
	BulkString   = '$'
	Array        = '*'
)

// Parser pour le protocole RESP
type Parser struct {
	reader *bufio.Reader
}

// NewParser crée un nouveau parser RESP
func NewParser(reader io.Reader) *Parser {
	return &Parser{
		reader: bufio.NewReader(reader),
	}
}

// ParseCommand parse une commande RESP complète
func (p *Parser) ParseCommand() ([]string, error) {
	// Lecture du premier caractère pour déterminer le type
	typeByte, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch typeByte {
	case Array:
		return p.parseArray()
	default:
		return nil, fmt.Errorf("expected array, got %c", typeByte)
	}
}

// parseArray parse un array RESP (format des commandes)
func (p *Parser) parseArray() ([]string, error) {
	// Lecture du nombre d'éléments
	lengthStr, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("failed to read array length: %v", err)
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %s", lengthStr)
	}

	if length <= 0 {
		return []string{}, nil
	}

	// Lecture de chaque élément
	elements := make([]string, length)
	for i := 0; i < length; i++ {
		element, err := p.parseBulkString()
		if err != nil {
			return nil, fmt.Errorf("failed to parse element %d: %v", i, err)
		}
		elements[i] = element
	}

	return elements, nil
}

// parseBulkString parse une bulk string RESP
func (p *Parser) parseBulkString() (string, error) {
	// Lecture du type (doit être $)
	typeByte, err := p.reader.ReadByte()
	if err != nil {
		return "", fmt.Errorf("failed to read bulk string type: %v", err)
	}

	if typeByte != BulkString {
		return "", fmt.Errorf("expected bulk string, got %c", typeByte)
	}

	// Lecture de la longueur
	lengthStr, err := p.readLine()
	if err != nil {
		return "", fmt.Errorf("failed to read bulk string length: %v", err)
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("invalid bulk string length: %s", lengthStr)
	}

	// Cas spécial : bulk string null
	if length == -1 {
		return "", nil
	}

	if length < 0 {
		return "", fmt.Errorf("invalid bulk string length: %d", length)
	}

	// Lecture du contenu
	content := make([]byte, length)
	_, err = io.ReadFull(p.reader, content)
	if err != nil {
		return "", fmt.Errorf("failed to read bulk string content: %v", err)
	}

	// Lecture du CRLF final
	crlf := make([]byte, 2)
	_, err = io.ReadFull(p.reader, crlf)
	if err != nil {
		return "", fmt.Errorf("failed to read CRLF after bulk string: %v", err)
	}

	if crlf[0] != '\r' || crlf[1] != '\n' {
		return "", fmt.Errorf("expected CRLF, got %v", crlf)
	}

	return string(content), nil
}

// readLine lit une ligne complète (jusqu'au CRLF)
func (p *Parser) readLine() (string, error) {
	var result []byte

	for {
		b, err := p.reader.ReadByte()
		if err != nil {
			return "", err
		}

		if b == '\r' {
			// Lire le \n suivant
			next, err := p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			if next != '\n' {
				return "", fmt.Errorf("expected \\n after \\r, got %c", next)
			}
			break
		}

		result = append(result, b)
	}

	return string(result), nil
}

// Encoder pour les réponses RESP
type Encoder struct {
	writer io.Writer
}

// NewEncoder crée un nouveau encoder RESP
func NewEncoder(writer io.Writer) *Encoder {
	return &Encoder{writer: writer}
}

// WriteSimpleString écrit une simple string (+OK)
func (e *Encoder) WriteSimpleString(s string) error {
	_, err := fmt.Fprintf(e.writer, "+%s\r\n", s)
	return err
}

// WriteError écrit une erreur (-ERR message)
func (e *Encoder) WriteError(message string) error {
	_, err := fmt.Fprintf(e.writer, "-%s\r\n", message)
	return err
}

// WriteInteger écrit un entier (:123)
func (e *Encoder) WriteInteger(i int64) error {
	_, err := fmt.Fprintf(e.writer, ":%d\r\n", i)
	return err
}

// WriteBulkString écrit une bulk string ($5\r\nhello\r\n)
func (e *Encoder) WriteBulkString(s string) error {
	_, err := fmt.Fprintf(e.writer, "$%d\r\n%s\r\n", len(s), s)
	return err
}

// WriteNullBulkString écrit une bulk string null ($-1\r\n)
func (e *Encoder) WriteNullBulkString() error {
	_, err := fmt.Fprintf(e.writer, "$-1\r\n")
	return err
}

// WriteArray écrit un array (*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n)
func (e *Encoder) WriteArray(elements []string) error {
	if _, err := fmt.Fprintf(e.writer, "*%d\r\n", len(elements)); err != nil {
		return err
	}

	for _, element := range elements {
		if err := e.WriteBulkString(element); err != nil {
			return err
		}
	}

	return nil
}
