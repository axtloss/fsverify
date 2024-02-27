#include "raylib.h"
#include "parseBVG.h"
#include "extString.h"
#include <stdbool.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>

int main(void) {
  FILE *bvgfile = readFile("./test.bvg");
  char *line = NULL;
  size_t len = 0;
  ssize_t nread = 0;
  ssize_t totallinesize = 0;
  ssize_t funcCount = 0;
  int inComment = 0;
  char *funcline = strdup("");
  char **funcs = malloc(1*sizeof(char));
  
  while ((nread = getline(&line, &len, bvgfile)) != -1) {
    if (strstr(line, "/*")) {
      inComment += 1;
      continue;
    }
    if (strstr(line, "*/")) {
      inComment -= 1;
      continue;
    }
    if (inComment != 0 || strstr(line, "//") || line[0] == '\n')
      continue;
    void *newfuncline = realloc(funcline, sizeof(char)*(strlen(funcline)+strlen(line)+1));
    if (newfuncline)
      funcline = newfuncline;
    else
      exit(2);
    totallinesize=totallinesize+nread;
    sprintf(funcline, "%s%s", funcline, line);
    if (line[nread-2] == ')') {
      funcCount += 1;
      void *newfuncs = realloc(funcs, sizeof(char)*(sizeof(funcs)+1+strlen(funcline)*2));
      if (newfuncs)
	funcs = newfuncs;
      else
	exit(2);
      funcs[funcCount-1]=strdup(funcline);
      totallinesize = 0;
      free(funcline);
      funcline = strdup("");
    }
  }
  free(line);

  for (int i = 0; i<funcCount; i++) {
    printf("%s", funcs[i]);
  }

  fclose(bvgfile);

  char *call = strdup(multiToSingle(funcs[0]));
  char *args[2];
  call=call+strlen("IMG (");
  char *callTrim = trim(call);
  callTrim[strlen(callTrim)-1]='\0';
  collectArgs(args, callTrim, 2);
  BVGIMG *imgsize = BVGParseIMG(args);

  InitWindow (imgsize->width, imgsize->height, ":3");


  while (!WindowShouldClose ()) {
    
    BeginDrawing ();
    ClearBackground (RAYWHITE);

    for (int i = 0; i<funcCount; i++) {
      matchFunctionCall(multiToSingle(funcs[i]));
    }

    char *text = malloc(strlen("100")*100);
    sprintf(text, "%d", GetFPS());
    DrawText(text, 2, 2, 20, MAROON);
    EndDrawing ();
  }

  CloseWindow ();
  
  return 0;
}
