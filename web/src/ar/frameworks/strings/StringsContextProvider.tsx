/*
 * Copyright 2024 Harness, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from 'react'
import languageLoader from './languageLoader'

import { StringsContext, StringsContextValue } from './StringsContext'

export interface StringsContextProviderProps extends Pick<StringsContextValue, 'getString'> {
  children: React.ReactNode
  initialStrings?: Record<string, any> // temp prop for backward compatibility
}

export function StringsContextProvider(props: StringsContextProviderProps): React.ReactElement {
  return (
    <StringsContext.Provider
      value={{
        data: {
          ...props.initialStrings,
          ...(languageLoader() as any)
        },
        getString: props.getString
      }}>
      {props.children}
    </StringsContext.Provider>
  )
}
