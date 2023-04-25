package main

import (
	"fmt"
	"math/rand"
	"os"
)

// Naive Approach
//
// Limitations:
// - Truncates file before updating. What if the file needs to read concurrently?
// - Write data to files is not atomic, depending on size of the write. Concurrent readers might get incomplete data.
// - When is the data actually persisted to disk? The data is probably still in the OS's page cache after the write
//   syscall returns. What is the state of the file when the system crashes or reboots?
func SaveData1(path string, data []byte) error {
  fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664);
  if err != nil {
    return err
  }
  defer fp.Close()

  _, err = fp.Write(data)
  return err
}

// Atomic Renaming
//
// To address the issue of concurrent readers getting incomplete data, we can write to a temporary file and then
// rename the file to the final destination. Renaming a file is an atomic operation on most filesystems.
//
// This is still problematic because it doesn't control when the data is persisted to the disk, and the metadata
// may be persisted to the disk before the data, potentially corrupting the file after a system crash
func SaveDate2(path string, data []byte) error {
  tmp := fmt.Sprintf("%s.tmp.%d", path, randomInt(100))
  fp, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664);
  if err != nil {
    return err
  }
  defer fp.Close()

  _, err = fp.Write(data)
  if err != nil {
    os.Remove(tmp)
    return err
  }

  return os.Rename(tmp, path)
}

func randomInt(max int) int {
  return rand.Intn(max)
}

// fsync
//
// To fix the problem address in SaveData2, we must flush the data to disk before renaming the file
// The Linux syscall for this is "fsync"
//
// This is still not a complete solution. The data has been flushed to disk, but what about the metadata?
// Should this also call fsync on the directory containing the file?
//
// All of these concerns combined make this a very complicated solution. This is why database are preferred
// over files for persisting data to the disk
func SaveData3(path string, data []byte) error {
  tmp := fmt.Sprintf("%s.tmp.%d", path, randomInt(100))
  fp, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664);
  if err != nil {
    return err
  }
  defer fp.Close()

  _, err = fp.Write(data)
  if err != nil {
    os.Remove(tmp)
    return err
  }

  err = fp.Sync()
  if err != nil {
    os.Remove(tmp)
    return err
  }

  return os.Rename(tmp, path)
}

// Append-Only Logs
//
// In some use cases, it makes sense to persis data using an append-only log
// 
// Append-only logs are nice because it does not modigy the existing data, nor does it deal with the rename
// operation, making it more resistant to corruption. Logs alone are not enough to build a database
// 
// 1. Database uses additional "indexes" to query the data efficiently. There are only brute-force ways to
//    query a bunch of records of arbitrary order
// 2. How do logs handle deleted data? They cannot grow forever
func LogCreate(path string) (*os.File, error) {
  return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
}

func LogAppend(fp *os.File, line string) error {
  buf := []byte(line)
  buf = append(buf, '\n')

  _, err := fp.Write(buf)
  if err != nil {
    return err
  }

  return fp.Sync()
}

func main() {
  path := "/Users/charlie/github.com/charlieroth/byodb/byodb.db"
  var data []byte = []byte{0, 1, 2, 3}

  err := SaveData1(path, data)
  if err != nil {
    fmt.Println("Failed to save data")
    os.Exit(1)
  }
}
