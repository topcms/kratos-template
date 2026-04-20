package server

import (
	"encoding/json"
	"net/http"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	infralogging "github.com/topcms/kratos-infra/middleware/logging"
)

type httpEnvelope struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func encodeHTTPResponse(w http.ResponseWriter, r *http.Request, v any) error {
	setResponseIDHeader(w, r)
	return writeHTTPEnvelope(w, r, http.StatusOK, httpEnvelope{
		Code: 0,
		Msg:  "ok",
		Data: v,
	})
}

func encodeHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	setResponseIDHeader(w, r)
	se := kerrors.FromError(err)
	statusCode := int(se.Code)
	if statusCode < http.StatusContinue || statusCode > 599 {
		statusCode = http.StatusInternalServerError
	}

	msg := se.Message
	if msg == "" {
		msg = se.Reason
	}
	if msg == "" {
		msg = http.StatusText(statusCode)
	}

	_ = writeHTTPEnvelope(w, r, statusCode, httpEnvelope{
		Code: se.Code,
		Msg:  msg,
		Data: nil,
	})
}

func writeHTTPEnvelope(w http.ResponseWriter, r *http.Request, statusCode int, body httpEnvelope) error {
	codec, _ := kratoshttp.CodecForRequest(r, "Accept")
	payload, err := codec.Marshal(body)
	contentType := "application/" + codec.Name()
	if err != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
		contentType = "application/json"
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	_, err = w.Write(payload)
	return err
}

func setResponseIDHeader(w http.ResponseWriter, r *http.Request) {
	responseID := infralogging.TraceIDFromContext(r.Context())
	if responseID == "" {
		responseID = r.Header.Get("X-Request-Id")
	}
	if responseID == "" {
		return
	}
	w.Header().Set("X-Request-Id", responseID)
}
