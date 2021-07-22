package locker

type EmptyLock struct {

}

func (l* EmptyLock) Lock()  {
	
}

func (l* EmptyLock) Unlock()  {

}
