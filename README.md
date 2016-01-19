# TryMutex

TryMutex is another synchronization primitive that additionaly to standard Lock and Unlock 
provides TryLock and TryLockTimeout methods.

  1. TryLock - tries to acquire lock returning true on success or false if failed.
  2. TryLockTimeout - tries to acquire a lock during specified time interval return false if time is out.

# NamedMutex

Named mutex is a syncroniation primitive that acquires lock based on name.
Primary usecase for myself is a lazy instantiation of objects based on name
that may take significant amount of time.


The following set of methods is available:
  
  1. Lock(name) - acquire lock by name.
  2. Unlock(name) - release lock by name.
  3. TryLock(name) - return true if lock acquired, otherwise false.
  4. TryLockTimeout(name, timeout) - tries to acquire lock for a certain amount of time. Returns false if timeouts.
  
