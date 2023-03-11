# nurikabe
Nurikabe solver in Go

## Usage

Problems are specified in a text file with a single character for each cell. Island sizes are specified with numerals for sizes 0-9, lowercase letters for sizes 10-35, or uppercase letters *through the letter W* for sizes 36-58, and empty cells with spaces or underscores. Partially solved puzzles can be specified with an X for a wall/painted cell and a dot (.) for a clear cell.

```
$ cat p3.txt
  7       
     7    
          
          
     2 5  
    4    6
      8   
    7     
          
          
   8      
2   2    3
 3        
    4     
$ go run main.go p3.txt
[===================] 140/140 (Stripping possibilities)
..7X..X.X.
.XXX.7X.X.
.X...XX.X.
.XXXX.X.X.
.X..X2X5X.
XXX.4XXXX6
X.XXXX8.XX
X...7X....
X.XXXXXXX.
X.X.....X.
.XX8XXX.XX
2X.X2.X.X3
X3.XXXXXX.
XXXX4...X.

Total duration: 0.9731
```
