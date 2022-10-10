package writer

import (
	"bytes"
	"net/http"

	"github.com/darkweak/go-esi/esi"
)

type Writer struct {
	buf       *bytes.Buffer
	rw        http.ResponseWriter
	Rq        *http.Request
	AsyncBuf  []chan []byte
	Done      chan bool
	flushed   bool
	Iteration int
}

func NewWriter(buf *bytes.Buffer, rw http.ResponseWriter, rq *http.Request) *Writer {
	return &Writer{
		buf:      buf,
		Rq:       rq,
		rw:       rw,
		AsyncBuf: make([]chan []byte, 0),
		Done:     make(chan bool),
	}
}

// Header implements http.ResponseWriter
func (w *Writer) Header() http.Header {
	return w.rw.Header()
}

// WriteHeader implements http.ResponseWriter
func (w *Writer) WriteHeader(statusCode int) {
	if statusCode == 0 {
		w.rw.WriteHeader(http.StatusOK)
	}
}

// Flush implements http.Flusher
func (w *Writer) Flush() {
	if !w.flushed {
		w.rw.(http.Flusher).Flush()
		w.flushed = true
	}
}

// Write will write the response body
func (w *Writer) Write(b []byte) (int, error) {
	buf := append(w.buf.Bytes(), b...)
	w.buf.Reset()

	if esi.HasOpenedTags(buf) {
		position := 0
		for position < len(buf) {
			startPos, nextPos, t := esi.ReadToTag(buf[position:], position)

			if startPos != 0 {
				w.AsyncBuf = append(w.AsyncBuf, make(chan []byte))
				go func(tmpBuf []byte, i int, cw *Writer) {
					cw.AsyncBuf[i] <- tmpBuf
				}(buf[position:position+startPos], w.Iteration, w)
				w.Iteration++
			}

			if t == nil {
				break
			}

			closePosition := t.GetClosePosition(buf[position+startPos:])
			if closePosition == 0 {
				position += startPos
				break
			}

			position += nextPos
			w.AsyncBuf = append(w.AsyncBuf, make(chan []byte))
			go func(currentTag esi.Tag, tmpBuf []byte, cw *Writer, Iteration int) {
				p, _ := currentTag.Process(tmpBuf, cw.Rq)
				cw.AsyncBuf[Iteration] <- p
			}(t, buf[position:(position-nextPos)+startPos+closePosition], w, w.Iteration)
			position += startPos + closePosition - nextPos
			w.Iteration++
		}
		w.buf.Write(buf[position:])
		return len(b), nil
	}

	w.AsyncBuf = append(w.AsyncBuf, make(chan []byte))
	w.AsyncBuf[w.Iteration] <- buf
	w.Iteration++
	return len(b), nil
}

var _ http.ResponseWriter = (*Writer)(nil)
