#include "raylib.h"
#include "parseBVG.h"
#include "extString.h"
#include <stdbool.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>

int getFuncs(char *file, char ***ret) {
  FILE *bvgfile = readFile(file);
  char *line = NULL;
  size_t len = 0;
  ssize_t nread = 0;
  ssize_t totallinesize = 0;
  ssize_t funcCount = 0;
  int inComment = 0;
  int newFuncsMem = sizeof(char)*1;
  char *funcline = strdup("");
  char **funcs = malloc(sizeof(char)*1);
  
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
    if (line[nread-2] == ')' || line[nread-1] == ')') {
      funcCount += 1;
      newFuncsMem = sizeof(char)*(sizeof(funcs)+sizeof(funcline)*funcCount);
      void *newfuncs = realloc(funcs, newFuncsMem);
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
  free(funcline);
  free(line);
  fclose(bvgfile);


  *ret = funcs;  
  return funcCount;  
}

int main(int argc, char **argv) {

  char **funcs;
  int funcCount = getFuncs(argv[1], &funcs);

  char *call = multiToSingle(funcs[0]);
  char *args[2];
  call=call+strlen("IMG (");
  char *callTrim = trim(call);
  callTrim[strlen(callTrim)-1]='\0';
  collectArgs(args, callTrim, 2);
  BVGIMG *imgsize = BVGParseIMG(args);

  InitWindow (imgsize->width, imgsize->height, ":3");

  free(imgsize);
  free(call-strlen("IMG ("));
  free(callTrim);
  for (int i = 0; i<funcCount; i++)
    free(funcs[i]);
  free(funcs);

  while (!WindowShouldClose ()) {
    char **funcs;
    int funcCount = getFuncs(argv[1], &funcs);
    free(funcs[0]); // Dont need you
    BeginDrawing ();
    ClearBackground (RAYWHITE);

    // i = 1 since the first item is always IMG
    for (int i = 1; i<funcCount; i++) {
      char *single = multiToSingle(funcs[i]);
      matchFunctionCall(single);
      free(single);
      free(funcs[i]);
    }

    char *text = malloc(strlen("100")*100);
    sprintf(text, "%d", GetFPS());
    DrawText(text, 2, 2, 20, MAROON);
    free(text);
    EndDrawing ();
    free(funcs);
  }

  CloseWindow ();
   
  return 0;
}
