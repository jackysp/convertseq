# convertseq

TiCDC doesn't support to sync sequences, this tool is used to sync sequences to the downstream TiDB cluster by convert it to a metadata table. When you want to restore the sequences, you can use this tool to restore the sequences from the metadata table.

Thanks [@Damon-Guo](https://github.com/Damon-Guo) for the idea and the implementation of the sync and restore part.

## Usage

```
➜  convertseq ./convertseq --help
Usage of ./convertseq:
  -mode string
    	Mode of operation: sync or restore
  -restoreIP string
    	Restore IP address (default "127.0.0.1")
  -restorePasswd string
    	Restore password (default "admin")
  -restorePort int
    	Restore port (default 4000)
  -restoreUser string
    	Restore user (default "admin")
  -restoreWorkers int
    	Number of workers for restore operation (default 5)
  -syncIP string
    	Sync IP address (default "127.0.0.1")
  -syncInterval duration
    	Sync interval (default 5s)
  -syncPasswd string
    	Sync password (default "admin")
  -syncPort int
    	Sync port (default 4000)
  -syncUser string
    	Sync user (default "admin")
```

## Example

```
➜  convertseq ./convertseq -mode sync -syncIP 127.0.0.1 -syncPort 4000 -syncUser root -syncPasswd ''               
Next row ID for sequence test.s: 1003
Next row ID for sequence test.s: 1003
Next row ID for sequence test.s: 1003
^C
➜  convertseq ./convertseq -mode restore -restoreIP 127.0.0.1 -restorePort 4000 -restoreUser root -restorePasswd '' -restoreWorkers 1
Dropping old sequences...
Restoring sequences...
Setting sequences...
```