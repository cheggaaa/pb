package pb

import "time"

type Progressable interface {
	Total() int64
	Value() int64
	Finished() bool
}

func RegisterProgressable(pr Progressable, removeFunc func(*ProgressBar)) *ProgressBar {
	pb := new(ProgressBar)
	go progressWorker(pr, pb, removeFunc)
	return pb
}

func progressWorker(pr Progressable, pb *ProgressBar, removeFunc func(*ProgressBar)) {
	for ; !pr.Finished(); time.Sleep(time.Second) {
		if !pb.IsStarted() {
			continue
		}
		_ = pb.SetCurrent(pr.Value()).SetTotal(pr.Total())
	}
	removeFunc(pb)
	pb.Finish()
}
