package interfacesize

type Store interface { // want "interface has 6 methods"
	Get() error
	Put() error
	Delete() error
	List() error
	Count() error
	Close() error
}

type small interface {
	Get() error
	Put() error
	Delete() error
	List() error
	Count() error
	Close() error
}

type Reader interface {
	Read() error
}
