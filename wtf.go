package frozen

type wtfError string

const wtf wtfError = "should never be called!"

func (e wtfError) Error() string {
	return string(e)
}
