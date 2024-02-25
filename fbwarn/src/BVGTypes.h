#include <raylib.h>
#include <stdbool.h>
typedef struct BVGRectangle {
  Rectangle rayrectangle;
  Color color;
  bool fill;
  float lineThickness;
} BVGRectangle;
