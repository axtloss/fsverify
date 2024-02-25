#include <stdio.h>
#include <stdlib.h>
#include "raylib.h"
#include "BVGTypes.h"

void BVGDrawRectangle(BVGRectangle*);
FILE *readFile(char*);
char *multiToSingle(char*);
void matchFunctionCall(char *);
BVGRectangle *BVGParseRectangle(char*[7]);
