# Refactoring with Go AST

Demonstrates how to use Go AST traversal to refactor a code base. In this particular case, we are adding a `context.Context` parameter to all methods defined on an interface. This is a contrived example derived from a real world use case in a very large code base that I worked on.

## Usage

```sh
cd refactor
go run main.go
```

The above will rewrite `main.go` in the `example` directory.

```diff
diff --git a/example/main.go b/example/main.go
index 3cb8d3b..2f7b048 100644
--- a/example/main.go
+++ b/example/main.go
@@ -1,6 +1,7 @@
 package main

 import (
+	"context"
 	"fmt"

 	"github.com/charithe/go-ast-refactoring/example/example"
@@ -10,6 +11,6 @@ func main() {
 	wc := example.WibbleClient{}
 	wcw := example.WibbleClientWrapper{WibbleClient: wc}

-	fmt.Printf("WibbleClient: %d\n", wc.Wibble(10))
-	fmt.Printf("WibbleClientWrapper: %d\n", wcw.Wibble(10))
+	fmt.Printf("WibbleClient: %d\n", wc.Wibble(context.Background(), 10))
+	fmt.Printf("WibbleClientWrapper: %d\n", wcw.Wibble(context.Background(), 10))
 }
```

