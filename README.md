# convertseq

TiCDC doesn't support to sync sequences, this tool is used to sync sequences to the downstream TiDB cluster by convert it to a metadata table. When you want to restore the sequences, you can use this tool to restore the sequences from the metadata table.

Thanks [@Damon-Guo](https://github.com/Damon-Guo) for the idea of the sync and restore part.

## Usage

```shell
➜  convertseq ./convertseq --help
Usage of ./convertseq:
  -logFilePath string
    	Path to error log file (default "error.log")
  -mode string
    	Mode of operation: sync or restore
  -restoreIP string
    	Restore IP address (default "127.0.0.1")
  -restorePasswd string
    	Restore password
  -restorePort int
    	Restore port (default 4000)
  -restoreSchema string
    	Restore Schema (default "test")
  -restoreUser string
    	Restore user (default "root")
  -restoreWorkers int
    	Number of workers for restore operation (default 5)
  -syncIP string
    	Sync IP address (default "127.0.0.1")
  -syncInterval duration
    	Sync interval (default 5s)
  -syncPasswd string
    	Sync password
  -syncPort int
    	Sync port (default 4000)
  -syncSchema string
    	Sync Schema (default "test")
  -syncUser string
    	Sync user (default "root")
```

## Example

```shell
➜  convertseq ./convertseq -mode sync -syncIP 127.0.0.1 -syncPort 4000 -syncUser root -syncPasswd ''               
All sequences updated at 2024-11-04 17:26:11.
All sequences updated at 2024-11-04 17:26:20.
All sequences updated at 2024-11-04 17:26:30.

➜  convertseq ./convertseq -mode restore -restoreIP 127.0.0.1 -restorePort 4000 -restoreUser root -restorePasswd '' -restoreWorkers 1
Creating new sequences...
Recreating changed sequences...
Dropping no longer needed sequences...
```