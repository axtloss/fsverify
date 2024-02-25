#include <raylib.h>
#include <stdbool.h>

typedef struct BVGRectangle {
  Rectangle rayrectangle;
  bool fill;
  float lineThickness;
  Color color;
} BVGRectangle;

typedef struct BVGRoundedRectangle {
  BVGRectangle rectangle;
  float roundness;
  int segments;
} BVGRoundedRectangle;

typedef struct BVGCircle {
  int centerX;
  int centerY;
  float radius;
  bool drawSector;
  float startAngle;
  float endAngle;
  int segments;
  Color color;
} BVGCircle;

typedef struct BVGRing {
  int centerX;
  int centerY;
  float inRadius;
  float outRadius;
  float startAngle;
  float endAngle;
  int segmets;
  Color color;
} BVGRing;

typedef struct BVGEllipse {
  int centerX;
  int centerY;
  float horizontalRadius;
  float verticalRadius;
  bool fill;
  Color color;
} BVGEllipse;

typedef struct BVGTriangle {
  Vector2 corner1;
  Vector2 corner2;
  Vector2 corner3;
  bool fill;
  Color color;
} BVGTriangle;

typedef struct BVGText {
  char *text;
  int x;
  int y;
  int fontSize;
  Color color;
} BVGText;
