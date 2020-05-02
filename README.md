# CRC32 Calculator with a Moving Window

This implement a CRC calculator with a moving window.  i.e. it will return the CRC32 value for the last N bytes read.

i.e. if one is looking for a an N byte section of a file that matches a specific CRC32 value, one can read it in byte by byte.

However, calculating the table to do the window moving can get very computationaly expensive for bigger windows (imagine windows that are 100s of MB big).  While one can compute that table with threads (it would scale linearly as no locking is needed), its stil can take 10s of minutes on an 8 core 3Ghz CPU.  To solve that problem, this code enables one to serialize out the tables that one wants to keep to not require computing them over and over again.

see ```example/find.go``` for a usage example