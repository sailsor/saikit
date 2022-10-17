package taskpool

type IJob interface {
	Run()
}

type JobCall func()

func (j JobCall) Run() {
	j()
}
