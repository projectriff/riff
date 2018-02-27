[![Coverage](http://gocover.io/_badge/github.com/NodePrime/jsonpath)](http://gocover.io/github.com/NodePrime/jsonpath)
# jsonpath  
  
jsonpath is used to pull values out of a JSON document without unmarshalling the string into an object.  At the loss of post-parse random access and conversion to primitive types, you gain faster return speeds and lower memory utilization.  If the value you want is located near the start of the json, the evaluator will terminate after reaching and recording its destination.  
  
The evaluator can be initialized with several paths, so you can retrieve multiple sections of the document with just one scan.  Naturally, when all paths have been reached, the evaluator will early terminate.  
  
For each value returned by a path, you'll also get the keys & indexes needed to reach that value.  Use the `keys` flag to view this in the CLI.  The Go package will return an `[]interface{}` of length `n` with indexes `0 - (n-2)` being the keys and the value at index `n-1`.  
  
### CLI   
```shell
go get github.com/NodePrime/jsonpath/cli/jsonpath
cat yourData.json | jsonpath -k -p '$.Items[*].title+'
```

##### Usage  
```shell
-f, --file="": Path to json file  
-j, --json="": JSON text  
-k, --keys=false: Print keys & indexes that lead to value  
-p, --path=[]: One or more paths to target in JSON
```

  
### Go Package  
go get github.com/NodePrime/jsonpath  
 
```go
paths, err := jsonpath.ParsePaths(pathStrings ...string) {
```  

```go
eval, err := jsonpath.GetPathsInBytes(json []byte, paths) 
// OR
eval, err := jsonpath.GetPathsInReader(r io.Reader, paths)
```

then  
```go  
for {
	if result, ok := eval.Next(); ok {
		fmt.Println(result.Pretty(true)) // true -> show keys in pretty string
	} else {
		break
	}
}
if eval.Error != nil {
	return eval.Error
}
```  

`eval.Next()` will traverse JSON until another value is found.  This has the potential of traversing the entire JSON document in an attempt to find one.  If you prefer to have more control over traversing, use the `eval.Iterate()` method.  It will return after every scanned JSON token and return `([]*Result, bool)`.  This array will usually be empty, but occasionally contain results.  
     
### Path Syntax  
All paths start from the root node `$`.  Similar to getting properties in a JavaScript object, a period `.` or brackets `[ .. ]` are used.  
  
`$` = root  
`.` = property of  
`["abc"]` = quoted property name  
`*` = wildcard property name  
`[n]` = Nth index of array  
`[n:m]` = Nth index to m-1 index (same as Go slicing)  
`[n:]` = Nth index to end of array  
`[*]` = wildcard index of array  
`+` = get value at end of path  
`?(expression)` = where clause (expression can reference current json node with @)
  
expression   
`@` = current node  
example:  
`?(@.value > 10)`  
   
Example: 
```javascript
{  
	"Items":   
		[  
			{  
				"title": "A Midsummer Night's Dream",  
				"tags":[  
					"comedy",  
					"shakespeare",  
					"play"  
				]  
			},{  
				"title": "A Tale of Two Cities",  
				"tags":[  
					"french",  
					"revolution",  
					"london"  
				]  
			}  
		]  
} 
```
	
Example Paths:   
*CLI*  
```shell
jsonpath --file=example.json --path='$.Items[*].tags[*]+' --keys
```   
"Items"	0	"tags"	0	"comedy"  
"Items"	0	"tags"	1	"shakespeare"  
"Items"	0	"tags"	2	"play"  
"Items"	1	"tags"	0	"french"  
"Items"	1	"tags"	1	"revolution"  
"Items"	1	"tags"	2	"london"  
  
*Paths*  
`$.Items[*].title+`   
... "A Midsummer Night's Dream"   
... "A Tale of Two Cities"   
  
`$.Items[*].tags+`    
... ["comedy","shakespeare","play"]  
... ["french","revolution","london"]  
  
`$.Items[*].tags[*]+`  
... "comedy"  
... "shakespeare"  
... "play"  
... "french"  
... "revolution"  
...  "london"  
  
... = keys/indexes of path  
