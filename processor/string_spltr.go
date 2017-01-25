package processor

type StringSplitter struct {
	In    chan string
	Out1  chan string
	Out2  chan string
	Error chan error
}

func NewStringSplitter(errChan chan error) *StringSplitter {
	return &StringSplitter{
		In:    make(chan string),
		Out1:  make(chan string),
		Out2:  make(chan string),
		Error: errChan,
	}
}

func (ss *StringSplitter) Run() {
	defer close(ss.Out1)
	defer close(ss.Out2)

	for str := range ss.In {
		ss.Out1 <- str
		ss.Out2 <- str
	}
}
