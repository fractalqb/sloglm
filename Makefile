.PHONY: clean

opts:
	go build -gcflags "-m -l" # go tool compile --help

compare: DefaultHeaderFast.prof AttrLookup.prof
	go tool pprof -http :6060 sloglm.test DefaultHeaderFast.prof &
	go tool pprof -http :6061 sloglm.test AttrLookup.prof &

kill:
	ps ax | fgrep "go tool pprof" | awk '{print $$1}' | xargs kill

fast: DefaultHeaderFast.prof
	go tool pprof -http :6060 sloglm.test DefaultHeaderFast.prof

attlup: AttrLookup.prof
	go tool pprof -http :6061 sloglm.test $<

clean:
	rm -f *.prof

%.prof:
	go test -cpuprofile $@ -bench=$*

mem.prof:
	go test -memprofile $@ -bench=$*
