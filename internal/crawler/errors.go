package crawler

import "errors"

var (
    ErrTimeout        = errors.New("request timeout")
    ErrInvalidLink    = errors.New("invalid link")
    ErrRequestFailed   = errors.New("request failed")
    ErrInvalidScheme  = errors.New("invalid scheme")
    ErrExternalDomain = errors.New("external domain")
    ErrNonHTMLResource = errors.New("non-HTML resource")
    ErrMaxDepthReached = errors.New("maximum depth reached")
    ErrRobotsDisallowed = errors.New("disallowed by robots.txt")
)