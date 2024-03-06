#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include "raylib.h"
#include "BVGTypes.h"

FILE *readFile(char *path);
char *multiToSingle(char *s);
void matchFunctionCall(char *call, float locScale);
void collectArgs(char *res[], char *call, int n);
void orderArgs(char *res[], char *argv[], int n, char *knownArgs[]);
Color *parseColorFromHex(char *hex);
bool parseBoolValue(char *hex);

// Shape functions
BVGIMG *BVGParseIMG(char *argv[2]);
BVGRectangle *BVGParseRectangle(char *argv[7]);
void BVGDrawRectangle(BVGRectangle *rectangle);
BVGRoundedRectangle *BVGParseRoundedRectangle(char *argv[9]);
void BVGDrawRoundedRectangle(BVGRoundedRectangle *rectangle);
BVGCircle *BVGParseCircle(char *argv[4]);
BVGCircle *BVGParseCircleSegment(char *argv[7]);
void BVGDrawCircle(BVGCircle *circle);
BVGRing *BVGParseRing(char *argv[8]);
void BVGDrawRing(BVGRing *ring);
BVGEllipse *BVGParseEllipse(char *argv[6]);
void BVGDrawEllipse(BVGEllipse *ellipse);
BVGTriangle *BVGParseTriangle(char *argv[8]);
void BVGDrawTriangle(BVGTriangle *triangle);
BVGText *BVGParseText(char *argv[5]);
void BVGDrawText(BVGText *text);
