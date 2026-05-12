package abbrev

type Doc struct{} // want "avoid abbreviation .*"

type Document struct{}

func ReqHandler() {} // want "avoid abbreviation .*"

func RequestHandler() {}
