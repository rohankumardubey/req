package value

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/andrewpillar/req/syntax"
)

// Response is the value for an HTTP response. This holds the underlying handle
// to the response.
type Response struct {
	*http.Response
}

// Select will return the value of the field with the given name.
func (r Response) Select(val Value) (Value, error) {
	name, err := ToName(val)

	if err != nil {
		return nil, err
	}

	switch name.Value {
	case "Status":
		return String{Value: r.Status}, nil
	case "StatusCode":
		return Int{Value: int64(r.StatusCode)}, nil
	case "Header":
		pairs := make(map[string]Value)
		order := make([]string, 0, len(r.Header))

		for k, v := range r.Header {
			order = append(order, k)
			pairs[k] = String{Value: v[0]}
		}
		return &Object{
			Order: order,
			Pairs: pairs,
		}, nil
	case "Body":
		if r.Body == nil {
			return &stream{}, nil
		}

		rc, rc2 := copyrc(r.Body)
		r.Body = rc

		b, _ := io.ReadAll(rc2)

		return NewStream(BufferStream(bytes.NewReader(b))), nil
	default:
		return nil, errors.New("type " + val.valueType().String() + " has no field " + name.Value)
	}
}

// String formats the response to a string. The formatted string will detail the
// pointer at which the underlying response handle exists.
func (r Response) String() string {
	return fmt.Sprintf("Response<addr=%p>", r.Response)
}

// Sprint formats the response into a string. This makes a copy of the response
// body so as to not deplete the original.
func (r Response) Sprint() string {
	if r.Response == nil {
		return ""
	}

	buf := bytes.NewBufferString(r.Proto + " " + r.Status + "\n")

	r.Header.Write(buf)

	if r.Body != nil {
		buf.WriteString("\n")

		rc, rc2 := copyrc(r.Body)

		r.Body = rc
		io.Copy(buf, rc2)
	}
	return buf.String()
}

func (r Response) valueType() valueType {
	return responseType
}

func (r Response) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, responseType)
}
