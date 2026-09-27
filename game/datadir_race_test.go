package game

import (
	"sync"
	"testing"
)

// The data-dir override is package state that tests swap while background
// goroutines resolve paths through it. Under -race this fails if the accesses
// are not synchronised (CI caught SetDataDirForTest's restore racing a reader).
func TestDataDirOverride_ConcurrentSwapAndRead(t *testing.T) {
	isolateAccountDir(t)
	a, b := t.TempDir(), t.TempDir()

	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				_ = rootDataDir()
				_ = SaveExists("autosave")
			}
		}
	}()
	for i := 0; i < 200; i++ {
		dir := a
		if i%2 == 1 {
			dir = b
		}
		SetDataDirForTest(dir)()
	}
	close(stop)
	wg.Wait()
}
