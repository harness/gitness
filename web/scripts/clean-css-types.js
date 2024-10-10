/*
 * Copyright 2023 Harness, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/* eslint-disable @typescript-eslint/no-var-requires, no-console */

/**
 * Since all the ".css.d.ts" files are generated automatically from webpack, developers might be
 * aware of its existence.
 *
 * Example: "MyAwesomeComponent.css" file will have a "MyAwesomeComponent.css.d.ts" file
 *
 * When a ".css" file is deleted, the corresponding ".css.d.ts" file must be deleted too, but this is
 * hard to do in a fast-paced development environment.
 *
 * The rationale here is that, since these files are generated automatically, they must be cleaned
 * automatically too.
 *
 * How do we do it?
 * We glob for all the ".css.d.ts" files and check if it has a corresponding ".css" file
 * If it doesn't, we delete that ".css.d.ts" file
 */

const fs = require('fs');
const glob = require('glob');

const files = glob.sync('src/**/*.css.d.ts');
console.log(`Found ${files.length} '.css.d.ts' files`);

let i = 0;

files.forEach(file => {
  // for every '.css' there will be a corresponding '.css.d.ts' file and vice versa
  const cssFile = file.replace('.d.ts', '');

  if (!fs.existsSync(cssFile)) {
    console.log(`Deleting "${file}" because corresponding "${cssFile}" does not exist`);
    fs.unlinkSync(file);
    i++;
  }
});

console.log(`Deleted total of ${i} '.css.d.ts' files`);
