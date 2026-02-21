package pdf

import _ "embed"

// Fonts contains the embedded Liberation Sans TTF font files
// used for ATS-compatible PDF generation.

//go:embed fonts/LiberationSans-Regular.ttf
var FontRegular []byte

//go:embed fonts/LiberationSans-Bold.ttf
var FontBold []byte

//go:embed fonts/LiberationSans-Italic.ttf
var FontItalic []byte

//go:embed fonts/LiberationSans-BoldItalic.ttf
var FontBoldItalic []byte
