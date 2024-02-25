#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include "raylib.h"
#include "BVGTypes.h"

FILE *readFile(char *);
char *multiToSingle(char *);
void matchFunctionCall(char *);
void collectArgs(char *[], char *, int);
void orderArgs(char *[], char *[], int, char *[]);
Color *parseColorFromHex(char *);
bool parseBoolValue(char *);

// Shape functions
BVGRectangle *BVGParseRectangle(char *[7]);
void BVGDrawRectangle(BVGRectangle*);
BVGRoundedRectangle *BVGParseRoundedRectangle(char *[9]);
void BVGDrawRoundedRectangle(BVGRoundedRectangle*);
BVGCircle *BVGParseCircle(char *[4]);
void BVGDrawCircle(BVGCircle);
