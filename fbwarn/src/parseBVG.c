#include <stdbool.h>
#include <stddef.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "BVGTypes.h"
#include "raylib.h"
#include "extString.h"

float scale;

void BVGDrawRectangle(BVGRectangle *rectangle) {
  if (rectangle->fill) {
    DrawRectangleRec(rectangle->rayrectangle, rectangle->color);
  } else {
    DrawRectangleLinesEx(rectangle->rayrectangle, rectangle->lineThickness, rectangle->color);
  }
  return;
}

void BVGDrawRoundedRectangle(BVGRoundedRectangle *rectangle) {
  if (rectangle->rectangle.fill) {
    DrawRectangleRounded(rectangle->rectangle.rayrectangle, rectangle->roundness, rectangle->segments, rectangle->rectangle.color);
  } else {
    DrawRectangleRoundedLines(rectangle->rectangle.rayrectangle, rectangle->roundness, rectangle->segments, rectangle->rectangle.lineThickness, rectangle->rectangle.color);
  }
  return;
}

void BVGDrawCircle(BVGCircle *circle) {
  Vector2 center = {circle->centerX, circle->centerY};
  if (circle->drawSector) {
    DrawCircleSector(center, circle->radius, circle->startAngle, circle->endAngle, circle->segments, circle->color);
  } else {
    DrawCircle(circle->centerX, circle->centerY, circle->radius, circle->color);
  }
  return;
}

void BVGDrawRing(BVGRing *ring) {
  Vector2 center = {ring->centerX, ring->centerY};
  DrawRing(center, ring->inRadius, ring->outRadius, ring->startAngle, ring->endAngle, ring->segmets, ring->color);
}

void BVGDrawEllipse(BVGEllipse *ellipse) {
  if (ellipse->fill) {
    DrawEllipse(ellipse->centerX, ellipse->centerY, ellipse->horizontalRadius, ellipse->verticalRadius, ellipse->color);
  } else {
    DrawEllipseLines(ellipse->centerX, ellipse->centerY, ellipse->horizontalRadius, ellipse->verticalRadius, ellipse->color);
  }
  return;
}

void BVGDrawTriangle(BVGTriangle *triangle) {
  if (triangle->fill) {
    DrawTriangle(triangle->corner1, triangle->corner2, triangle->corner3, triangle->color);
  } else {
    DrawTriangleLines(triangle->corner1, triangle->corner2, triangle->corner3, triangle->color);
  }
  return;
}

void BVGDrawText(BVGText *text) {
  DrawText(text->text, text->x, text->y, text->fontSize, text->color);
  return;
}

// reads a given file
FILE *readFile(char *path) {
  FILE *fp = fopen (path, "r");
  if (fp == NULL) {
    perror ("Failed to open file!");
    exit(1);
  }
  return fp;
}

/*
  Converts a multiline function
  call into a single line call
*/
char *multiToSingle(char *s) {
  char *midresult;
  char *notab;
  char *result;
  midresult = replaceStr(s, "\n", "");
  notab = replaceStr(midresult, "\t", "");
  result = replaceStr(midresult, ", ", ",");
  free(midresult);
  free(notab);
  return result;
}

/*
  Collects all arguments from a function call
  arguments are seperated with a ,
*/
void collectArgs(char *res[], char *call, int n) {
  char **arg;
  char *args[n];
  for (arg = args; (*arg = strsep(&call, ",")) != NULL;)
	if (**arg != '\0')
	  if (++arg >= &args[n])
 	    break;
  memcpy(res, args, sizeof(args));
}

/*
  Order the arguments in the order specified with knownArgs
  the res array contains only the values of each argument
*/
void orderArgs(char *res[], char *argv[], int n, char *knownArgs[]) {
  if (argv == NULL || knownArgs == NULL)
    exit(1);
  for(int i=0; i<n; i++) {
    if (argv[i] == NULL)
      continue;
    for(int j=0; j<n; j++) {
      if (strncmp(argv[i], knownArgs[j], strlen(knownArgs[j])) == 0) {
	res[j] = argv[i]+strlen(knownArgs[j])+1;
      }
    }
  }
}

/*
  Parses a color from a hex representation
  supports RRGGBB and RRGGBBAA
  assumes that only the color value is in the string
*/
Color *parseColorFromHex(char *hex) {
  Color *clr = malloc(sizeof(Color));
  int r,g,b,a = 0;
  if (strlen(hex) == 6) {
    sscanf(hex, "%02x%02x%02x", &r, &g, &b);
    clr->r=r; clr->g=g; clr->b=b; clr->a=255;
  } else {
    sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a);
    clr->r=r; clr->g=g; clr->b=b; clr->a=a;
  }
  return clr;
}

/*
  Converts a given char pointer to a corresponding boolean value
  expects the char to only contain true/false without any whitespace or similiar
  case is ignored
  If no match is found, true is returned by default
*/
bool parseBoolValue(char *value) {
  char *valueLower = strdup(strlwr(value));
  if (strcmp(valueLower, "false") == 0) {
    free(valueLower);
    return false;
  } else {
    free(valueLower);
    return true;
  }
}

BVGIMG *BVGParseIMG(char *argv[2]) {
  BVGIMG *result = malloc(sizeof(BVGIMG));
  size_t argN = 2;
  char *args[argN];
  char *knownArgs[2] = {"width", "height"};
  orderArgs(args, argv, argN, knownArgs);
  int width, height;
  sscanf(args[0], "%d", &width);
  sscanf(args[1], "%d", &height);
  result->width=width; result->height=height;
  return result;
}

/*
  Creates a BVGRectangle based on a BVG function call
*/
BVGRectangle *BVGParseRectangle(char *argv[7]) {
  BVGRectangle *result = malloc(sizeof(BVGRectangle));
  Rectangle *rectangle = malloc(sizeof(Rectangle));
  size_t argN = 7;
  char *args[argN];
  char *knownArgs[7] = {"x", "y", "width", "height", "color", "fill", "thickness"};
  int x, y, width, height, r, g, b, a;
  float thickness = 1.0;
  bool fill = true;
  Color *clr;
  
  orderArgs(args, argv, argN, knownArgs);
  sscanf(args[0], "%d", &x);
  sscanf(args[1], "%d", &y);
  sscanf(args[2], "%d", &width);
  sscanf(args[3], "%d", &height);
  args[4] = args[4]+2;
  args[4][strlen(args[4])-1] = '\0';
  clr = parseColorFromHex(args[4]);
  sscanf(args[6], "%fd", &thickness);
  fill = parseBoolValue(args[5]);
 
  rectangle->x=x*scale; rectangle->y=y*scale;
  rectangle->height=height*scale; rectangle->width=width*scale;
  result->rayrectangle=*rectangle;
  result->color=*clr;
  result->fill=fill;
  result->lineThickness=thickness*scale;

  free(clr);
  free(rectangle);
  return result;
}

/*
  Creates a BVGRoundedrectangle based on a BVG function call
*/
BVGRoundedRectangle *BVGParseRoundedRectangle(char *argv[9]) {
  BVGRoundedRectangle *result = malloc(sizeof(BVGRoundedRectangle));
  BVGRectangle *bvgrectangle;
  Rectangle *rectangle;
  size_t argN = 9;
  char *args[argN];
  char *knownArgs[9] = {"x", "y", "width", "height", "color", "fill", "thickness", "roundness", "segments"};
  float roundness;
  int segments;
  
  orderArgs(args, argv, argN, knownArgs);
  bvgrectangle = BVGParseRectangle(argv);
  result->rectangle = *bvgrectangle;
  sscanf(args[7], "%fd", &roundness);
  sscanf(args[8], "%d", &segments);

  result->roundness = roundness;
  result->segments = segments;
  
  free(bvgrectangle);
  return result;
}

BVGCircle *BVGParseCircle(char *argv[4]) {
  BVGCircle *result = malloc(sizeof(BVGCircle));
  size_t argN = 4;
  char *args[argN];
  char *knownArgs[4] = {"x", "y", "radius", "color"};
  int x, y;
  float radius;
  Color *clr;
  
  orderArgs(args, argv, argN, knownArgs);
  sscanf(args[0], "%d", &x);
  sscanf(args[1], "%d", &y);
  sscanf(args[2], "%fd", &radius);
  args[3] = args[3]+2;
  args[3][strlen(args[3])-1] = '\0';
  clr = parseColorFromHex(args[3]);

  result->drawSector=false;
  result->centerX=x*scale;
  result->centerY=y*scale;
  result->radius=radius*scale;
  result->color=*clr;
  
  free(clr);
  return result;
}

BVGCircle *BVGParseCircleSegment(char *argv[7]) {
  BVGCircle *result = malloc(sizeof(BVGCircle));
  size_t argN = 7;
  char *args[argN];
  char *knownArgs[7] = {"x", "y", "radius", "color", "startangle", "endangle", "segments"};
  int x, y, segments;
  float radius, startAngle, endAngle;
  Color *clr;
  
  orderArgs(args, argv, argN, knownArgs);
  sscanf(args[0], "%d", &x);
  sscanf(args[1], "%d", &y);
  sscanf(args[2], "%fd", &radius);
  sscanf(args[4], "%fd", &startAngle);
  sscanf(args[5], "%fd", &endAngle);
  sscanf(args[6], "%d", &segments);
  args[3] = args[3]+2;
  args[3][strlen(args[3])-1] = '\0';
  clr = parseColorFromHex(args[3]);

  result->drawSector=true;
  result->centerX=x*scale; result->centerY=y*scale;
  result->color=*clr;
  result->radius=radius*scale;
  result->segments=segments;
  result->startAngle=startAngle; result->endAngle=endAngle;
  
  free(clr);
  return result;
}

BVGRing *BVGParseRing(char *argv[8]) {
  BVGRing *result = malloc(sizeof(BVGRing));
  size_t argN = 8;
  char *args[argN];
  char *knownArgs[8] = {"x", "y", "innerradius", "outerradius", "startangle", "endangle", "segments", "color"};
  int x, y, segments;
  float innerRadius, outerRadius, startAngle, endAngle;
  Color *clr;
  
  orderArgs(args, argv, argN, knownArgs);
  sscanf(args[0], "%d", &x);
  sscanf(args[1], "%d", &y);
  sscanf(args[2], "%f", &innerRadius);
  sscanf(args[3], "%f", &outerRadius);
  sscanf(args[4], "%f", &startAngle);
  sscanf(args[5], "%f", &endAngle);
  sscanf(args[6], "%d", &segments);
  args[7] = args[7]+2;
  args[7][strlen(args[7])-1] = '\0';
  clr = parseColorFromHex(args[7]);
  

  result->centerX=x*scale; result->centerY=y*scale;
  result->inRadius=innerRadius*scale; result->outRadius=outerRadius*scale;
  result->startAngle=startAngle; result->endAngle=endAngle;
  result->segmets=segments;
  result->color=*clr;
  
  free(clr);
  return result;
}

BVGEllipse *BVGParseEllipse(char *argv[6]) {
  BVGEllipse *result = malloc(sizeof(BVGEllipse));
  size_t argN = 6;
  char *args[argN];
  char *knownArgs[6] = {"x", "y", "horizontalradius", "verticalradius", "fill", "color"};
  orderArgs(args, argv, argN, knownArgs);
  int x, y;
  float horizontalRadius, verticalRadius;
  bool fill;
  Color *clr;
  sscanf(args[0], "%d", &x);
  sscanf(args[1], "%d", &y);
  sscanf(args[2], "%f", &horizontalRadius);
  sscanf(args[3], "%f", &verticalRadius);
  fill = parseBoolValue(args[4]);
  args[5] = args[5]+2;
  args[5][strlen(args[5])-1] = '\0';
  clr = parseColorFromHex(args[5]);
  
  result->centerX=x*scale; result->centerY=y*scale;
  result->horizontalRadius=horizontalRadius*scale;
  result->verticalRadius=verticalRadius*scale;
  result->fill=fill;
  result->color=*clr;

  free(clr);
  return result;
}  

BVGTriangle *BVGParseTriangle(char *argv[8]) {
  BVGTriangle *result = malloc(sizeof(BVGTriangle));
  size_t argN = 8;
  char *args[argN];
  char *knownArgs[8] = {"x1", "y1", "x2", "y2", "x3", "y3", "fill", "color"};
  int x1, x2, x3;
  int y1, y2, y3;
  bool fill;
  Color *clr;
  
  orderArgs(args, argv, argN, knownArgs);
  sscanf(args[0], "%d", &x1);
  sscanf(args[1], "%d", &y1);
  sscanf(args[2], "%d", &x2);
  sscanf(args[3], "%d", &y2);
  sscanf(args[4], "%d", &x3);
  sscanf(args[5], "%d", &y3);
  fill = parseBoolValue(args[6]);
  args[7] = args[7]+2;
  args[7][strlen(args[7])-1] = '\0';
  clr = parseColorFromHex(args[7]);
  
  result->corner1 = (Vector2){x1*scale, y1*scale};
  result->corner2 = (Vector2){x2*scale, y2*scale};
  result->corner3 = (Vector2){x3*scale, y3*scale};
  result->fill = fill;
  result->color = *clr;

  free(clr);
  return result;
}

BVGText *BVGParseText(char *argv[5]) {
  BVGText *result = malloc(sizeof(BVGText));
  size_t argN = 5;
  char *args[argN];
  char *knownArgs[5] = {"text", "x", "y", "size", "color"};
  int x, y, size;
  Color *clr;
  
  orderArgs(args, argv, argN, knownArgs);
  args[0] = args[0]+1;
  args[0][strlen(args[0])-1] = '\0';
  sscanf(args[1], "%d", &x);
  sscanf(args[2], "%d", &y);
  sscanf(args[3], "%d", &size);
  args[4] = args[4]+2;
  args[4][strlen(args[4])-1] = '\0';
  clr = parseColorFromHex(args[4]);

  result->text=args[0]; result->fontSize=size*scale;
  result->x=x*scale; result->y=y*scale; 
  result->color=*clr;

  free(clr);
  return result;
}

/*
  Takes a BVG function call and calls the according C function
*/
void matchFunctionCall(char *call, float locScale) {
  scale = locScale;
  char *funcCall = strdup(call);
  char *funcName = strsep(&funcCall, "(");
  char *function = trim(strlwr(funcName));
  free(funcName);
  call[strlen(call)-1]='\0';
  if (strcmp(function, "rectangle") == 0) {
    char *argv[7];
    call = call+strlen("rectangle (");
    collectArgs(argv, call, 7);
    BVGRectangle *shape = BVGParseRectangle(argv);
    BVGDrawRectangle(shape);
    free(shape);
  } else if (strcmp(function, "roundedrectangle") == 0) {
    char *argv[9];
    call = call+strlen("roundedrectangle (");
    collectArgs(argv, call, 9);
    BVGRoundedRectangle *shape = BVGParseRoundedRectangle(argv);
    BVGDrawRoundedRectangle(shape);
    free(shape);
  } else if (strcmp(function, "circle") == 0) {
    char *argv[4];
    call = call+strlen("circle (");
    collectArgs(argv, call, 4);
    BVGCircle *shape = BVGParseCircle(argv);
    BVGDrawCircle(shape);
    free(shape);
  } else if (strcmp(function, "circlesegment") == 0) {
    char *argv[7];
    call = call+strlen("circlesegment (");
    collectArgs(argv, call, 7);
    BVGCircle *shape = BVGParseCircleSegment(argv);
    BVGDrawCircle(shape);
    free(shape);
  } else if (strcmp(function, "ring") == 0) {
    char *argv[8];
    call = call+strlen("ring (");
    collectArgs(argv, call, 8);
    BVGRing *shape = BVGParseRing(argv);
    BVGDrawRing(shape);
    free(shape);
  } else if (strcmp(function, "ellipse") == 0) {
    char *argv[6];
    call = call+strlen("ellipse (");
    collectArgs(argv, call, 6);
    BVGEllipse *shape = BVGParseEllipse(argv);
    BVGDrawEllipse(shape);
    free(shape);
  } else if (strcmp(function, "triangle") == 0) {
    char *argv[8];
    call = call+strlen("triangle (");
    collectArgs(argv, call, 8);
    BVGTriangle *shape = BVGParseTriangle(argv);
    BVGDrawTriangle(shape);
    free(shape);
  } else if (strcmp(function, "text") == 0) {
    char *argv[5];
    call = call+strlen("text (");
    collectArgs(argv, call, 5);
    BVGText *shape =  BVGParseText(argv);
    BVGDrawText(shape);
    free(shape);
  }
  free(function);
  return;
}
