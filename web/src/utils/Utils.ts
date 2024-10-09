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

import { IconName, Intent, IToaster, IToastProps, Position, Toaster } from '@blueprintjs/core'
import { get } from 'lodash-es'
import moment from 'moment'
import langMap from 'lang-map'
import type { EditorDidMount } from 'react-monaco-editor'
import type { editor } from 'monaco-editor'
import type { EditorView } from '@codemirror/view'
import type { FormikProps } from 'formik'
import type {
  TypesRuleViolations,
  TypesViolation,
  TypesCodeOwnerEvaluationEntry,
  RepoRepositoryOutput,
  TypesLabel,
  TypesLabelValue
} from 'services/code'

export enum ACCESS_MODES {
  VIEW,
  EDIT
}

export enum PullRequestSection {
  CONVERSATION = 'conversation',
  COMMITS = 'commits',
  FILES_CHANGED = 'changes',
  CHECKS = 'checks'
}

export enum FeatureType {
  COMINGSOON = 'comingSoon',
  RELEASED = 'released'
}

export enum REPO_EXPORT_STATE {
  FINISHED = 'finished',
  FAILED = 'failed',
  CANCELED = 'canceled',
  RUNNING = 'running',
  SCHEDULED = 'scheduled'
}

export const LIST_FETCHING_LIMIT = 20
export const DEFAULT_DATE_FORMAT = 'MM/DD/YYYY hh:mm a'
export const DEFAULT_BRANCH_NAME = 'main'
export const REGEX_VALID_REPO_NAME = /^[a-zA-Z_][0-9a-zA-Z-_.$]*$/
export const REGEX_VALID_USER_ID = /^[0-9a-zA-Z][0-9a-zA-Z-_.]*$/
export const SUGGESTED_BRANCH_NAMES = [DEFAULT_BRANCH_NAME, 'master']
export const FILE_SEPERATOR = '/'
export const INITIAL_ZOOM_LEVEL = 1
export const ZOOM_INC_DEC_LEVEL = 0.1

/** This utility shows a toaster without being bound to any component.
 * It's useful to show cross-page/component messages */
export function showToaster(message: string, props?: Partial<IToastProps>): IToaster {
  const toaster = Toaster.create({ position: Position.TOP })
  toaster.show({ message, intent: Intent.SUCCESS, ...props })
  return toaster
}

export function permissionProps(permResult: { disabled: boolean; tooltip: JSX.Element | string }, standalone: boolean) {
  const perm = standalone ? true : false
  return !perm ? permResult : undefined
}

export function generateAlphaNumericHash(length: number) {
  let result = ''
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  const charactersLength = characters.length

  for (let i = 0; i < length; i++) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength))
  }

  return result
}

export const dayAgoInMS = 86400000

export const getErrorMessage = (error: Unknown): string | undefined =>
  error ? get(error, 'data.error', get(error, 'data.message', get(error, 'message', error))) : undefined

export interface PageBrowserProps {
  page?: string
  state?: string
  tab?: string
}

export const extractInfoFromRuleViolationArr = (ruleViolationArr: TypesRuleViolations[]) => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const tempArray: any[] = ruleViolationArr?.flatMap(
    (item: { violations?: TypesViolation[] | null }) => item?.violations?.map(violation => violation.message) ?? []
  )
  const uniqueViolations = new Set(tempArray)
  const violationArr = [...uniqueViolations].map(violation => ({ violation: violation }))

  const checkIfBypassNotAllowed = ruleViolationArr.some(ruleViolation => ruleViolation.bypassed === false)

  return {
    uniqueViolations,
    checkIfBypassNotAllowed,
    violationArr
  }
}

export interface RenameDetails {
  commit_sha_after: string
  commit_sha_before: string
  new_path: string
  old_path: string
}

export interface LoginForm {
  username: string
  password: string
}

export interface RegisterForm {
  username: string
  password: string
  confirmPassword: string
  email: string
}

export interface FeatureData {
  typeText: string
  type: string
  title: string
  content: string
  link: string
}

export interface SourceCodeEditorProps {
  source: string
  language?: string
  lineNumbers?: boolean
  readOnly?: boolean
  highlightLines?: string // i.e: {1,3-4}, TODO: not yet supported
  height?: number | string
  autoHeight?: boolean
  wordWrap?: boolean
  onChange?: (value: string) => void
  schema?: Record<string, unknown>
  editorDidMount?: EditorDidMount
  editorOptions?: editor.IStandaloneEditorConstructionOptions
}

export type FieldCheck = {
  [key: string]: string
}

export const displayDateTime = (value: number): string | null => {
  return value ? moment.unix(value / 1000).format(DEFAULT_DATE_FORMAT) : null
}

export const timeDistance = (date1 = 0, date2 = 0, onlyHighestDenomination = false) => {
  let distance = Math.abs(date1 - date2)

  if (!distance) {
    return '0s'
  }

  const days = Math.floor(distance / (24 * 3600000))
  if (onlyHighestDenomination && days) {
    return days + 'd'
  }
  distance -= days * 24 * 3600000

  const hours = Math.floor(distance / 3600000)
  if (onlyHighestDenomination && hours) {
    return hours + 'h'
  }
  distance -= hours * 3600000

  const minutes = Math.floor(distance / 60000)
  if (onlyHighestDenomination && minutes) {
    return minutes + 'm'
  }
  distance -= minutes * 60000

  const seconds = Math.floor(distance / 1000)
  if (onlyHighestDenomination) {
    return seconds + 's'
  }

  return `${days ? days + 'd ' : ''}${hours ? hours + 'h ' : ''}${
    minutes ? minutes + 'm' : hours || days ? '0m' : ''
  } ${seconds}s`
}

export const timeDifferenceInMinutesAndSeconds = (date1 = 0, date2 = 0) => {
  let distance = Math.abs(date1 - date2)

  if (!distance) {
    return '00:00'
  }

  const minutes = Math.floor(distance / 60000)
  distance -= minutes * 60000

  const seconds = Math.floor(distance / 1000)

  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
}

const LOCALE = Intl.NumberFormat().resolvedOptions?.().locale || 'en-US'

/**
 * Format a timestamp to short format time (i.e: 7:41 AM)
 * @param timestamp Timestamp
 * @param timeStyle Optional DateTimeFormat's `timeStyle` option.
 */
export function formatTime(timestamp: number | string, timeStyle = 'short'): string {
  return timestamp
    ? new Intl.DateTimeFormat(LOCALE, {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore: TS built-in type for DateTimeFormat is not correct
        timeStyle
      }).format(new Date(timestamp))
    : ''
}

type FileDropCallback = (file: File) => void
//handle file drop in image upload
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const handleFileDrop = (event: any, callback: FileDropCallback): void => {
  event.preventDefault()

  const file = event?.dataTransfer?.files[0]
  if (file) {
    callback(file)
  }
}

type PasteCallback = (file: File) => void

// handle file paste in image upload
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const handlePaste = (event: { preventDefault: () => void; clipboardData: any }, callback: PasteCallback) => {
  event.preventDefault()
  const clipboardData = event.clipboardData
  const items = clipboardData.items

  if (items.length > 0) {
    const firstItem = items[0]
    if (firstItem.type.startsWith('image/') || firstItem.type.startsWith('video/')) {
      const blob = firstItem.getAsFile()
      callback(blob)
    }
  }
}

// find code owner request decision from given entries
export const findChangeReqDecisions = (
  entries: TypesCodeOwnerEvaluationEntry[] | null | undefined,
  decision: string
) => {
  if (entries === null || entries === undefined) {
    return []
  } else {
    return entries
      .map((entry: TypesCodeOwnerEvaluationEntry) => {
        // Filter the owner_evaluations for 'changereq' decisions
        const changeReqEvaluations = entry?.owner_evaluations?.filter(
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          (evaluation: any) => evaluation?.review_decision === decision
        )

        // If there are any 'changereq' decisions, return the entry along with them
        if (changeReqEvaluations && changeReqEvaluations?.length > 0) {
          return { ...entry, owner_evaluations: changeReqEvaluations }
        }
      }) // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .filter((entry: any) => entry !== null && entry !== undefined) // Filter out the null entries
  }
}

export const findWaitingDecisions = (entries: TypesCodeOwnerEvaluationEntry[] | null | undefined) => {
  if (entries === null || entries === undefined) {
    return []
  } else {
    return entries.filter((entry: TypesCodeOwnerEvaluationEntry) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const hasEmptyDecision = entry?.owner_evaluations?.some((evaluation: any) => evaluation?.review_decision === '')
      const hasNoApprovedDecision = !entry?.owner_evaluations?.some(
        evaluation => evaluation.review_decision === 'approved'
      )

      return hasEmptyDecision && hasNoApprovedDecision
    })
  }
}

/**
 * Format a timestamp to medium format date (i.e: Jan 1, 2021)
 * @param timestamp Timestamp
 * @param dateStyle Optional DateTimeFormat's `dateStyle` option.
 */
export function formatDate(timestamp: number | string, dateStyle = 'medium'): string {
  return timestamp
    ? new Intl.DateTimeFormat(LOCALE, {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore: TS built-in type for DateTimeFormat is not correct
        dateStyle
      }).format(new Date(timestamp))
    : ''
}

/**
 * Format a number with current locale.
 * @param num number
 * @returns Formatted string.
 */
export function formatNumber(num: number | bigint): string {
  return num ? new Intl.NumberFormat(LOCALE).format(num) : ''
}
export interface Violation {
  violation: string
}

export const rulesFormInitialPayload: RulesFormPayload = {
  name: '',
  desc: '',
  enable: true,
  target: '',
  targetDefault: false,
  targetList: [] as string[][],
  allRepoOwners: false,
  bypassList: [] as string[],
  requireMinReviewers: false,
  minReviewers: '',
  requireCodeOwner: false,
  requireNewChanges: false,
  reqResOfChanges: false,
  requireCommentResolution: false,
  requireStatusChecks: false,
  statusChecks: [] as string[],
  limitMergeStrategies: false,
  mergeCommit: false,
  squashMerge: false,
  rebaseMerge: false,
  autoDelete: false,
  blockBranchCreation: false,
  blockBranchDeletion: false,
  blockBranchUpdate: false,
  blockForcePush: false,
  requirePr: false,
  bypassSet: false,
  targetSet: false
}

export type RulesFormPayload = {
  name?: string
  desc?: string
  enable: boolean
  target?: string
  targetDefault?: boolean
  targetList: string[][]
  allRepoOwners?: boolean
  bypassList?: string[]
  requireMinReviewers: boolean
  minReviewers?: string | number
  requireCodeOwner?: boolean
  requireNewChanges?: boolean
  reqResOfChanges?: boolean
  requireCommentResolution?: boolean
  requireStatusChecks: boolean
  statusChecks: string[]
  limitMergeStrategies: boolean
  mergeCommit?: boolean
  squashMerge?: boolean
  rebaseMerge?: boolean
  autoDelete?: boolean
  blockBranchCreation?: boolean
  blockBranchDeletion?: boolean
  blockBranchUpdate?: boolean
  blockForcePush?: boolean
  requirePr?: boolean
  bypassSet: boolean
  targetSet: boolean
}

/**
 * Make any HTML element as a clickable button with keyboard accessibility
 * support (hit Enter/Space will trigger click event)
 */
export const ButtonRoleProps = {
  onKeyDown: (e: React.KeyboardEvent<HTMLElement>) => {
    if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar' || e.which === 13 || e.which === 32) {
      ;(e.target as unknown as { click: () => void })?.click?.()
    }
  },
  tabIndex: 0,
  role: 'button',
  style: { cursor: 'pointer ' }
}

export enum orderSortDate {
  ASC = 'asc',
  DESC = 'desc'
}

export enum CodeCommentState {
  ACTIVE = 'active',
  RESOLVED = 'resolved'
}

export enum FileSection {
  CONTENT = 'content',
  BLAME = 'blame',
  HISTORY = 'history'
}

export enum ApproveState {
  APPROVED = 'approved',
  CHANGEREQ = 'changereq',
  APPROVE = 'approve',
  OUTDATED = 'outdated'
}

export enum CheckStatus {
  PENDING = 'pending',
  RUNNING = 'running',
  SUCCESS = 'success',
  FAILURE = 'failure',
  ERROR = 'error',
  SKIPPED = 'skipped',
  KILLED = 'killed'
}

export enum PRCommentFilterType {
  SHOW_EVERYTHING = 'showEverything',
  ALL_COMMENTS = 'allComments',
  MY_COMMENTS = 'myComments',
  RESOLVED_COMMENTS = 'resolvedComments',
  UNRESOLVED_COMMENTS = 'unresolvedComments'
}

const MONACO_SUPPORTED_LANGUAGES = [
  'abap',
  'apex',
  'azcli',
  'bat',
  'bicep',
  'cameligo',
  'clojure',
  'coffee',
  'cpp',
  'csharp',
  'csp',
  'css',
  'cypher',
  'dart',
  'dockerfile',
  'ecl',
  'elixir',
  'flow9',
  'freemarker2',
  'fsharp',
  'go',
  'graphql',
  'handlebars',
  'hcl',
  'html',
  'ini',
  'java',
  'javascript',
  'json',
  'julia',
  'kotlin',
  'less',
  'lexon',
  'liquid',
  'lua',
  'm3',
  'markdown',
  'mips',
  'msdax',
  'mysql',
  'objective-c',
  'pascal',
  'pascaligo',
  'perl',
  'pgsql',
  'php',
  'pla',
  'postiats',
  'powerquery',
  'powershell',
  'protobuf',
  'pug',
  'python',
  'qsharp',
  'r',
  'razor',
  'redis',
  'redshift',
  'restructuredtext',
  'ruby',
  'rust',
  'sb',
  'scala',
  'scheme',
  'scss',
  'shell',
  'solidity',
  'sophia',
  'sparql',
  'sql',
  'st',
  'swift',
  'systemverilog',
  'tcl',
  'twig',
  'typescript',
  'vb',
  'wgsl',
  'xml',
  'yaml'
]

// Some of the below languages are mapped to Monaco's built-in supported languages
// due to their similarity. We'll still need to get native support for them at
// some point.
const EXTENSION_TO_LANG: Record<string, string> = {
  alpine: 'dockerfile',
  bazel: 'python',
  cc: 'cpp',
  cs: 'csharp',
  env: 'shell',
  gitignore: 'shell',
  jsx: 'typescript',
  makefile: 'shell',
  toml: 'ini',
  tsx: 'typescript',
  tf: 'hcl',
  tfvars: 'hcl',
  workspace: 'python',
  tfstate: 'hcl',
  ipynb: 'json'
}

export const PLAIN_TEXT = 'plaintext'

export const filenameToLanguage = (name?: string): string | undefined => {
  const extension = (name?.split('.').pop() || '').toLowerCase()
  const lang = MONACO_SUPPORTED_LANGUAGES.find(l => l === extension) || EXTENSION_TO_LANG[extension]

  if (lang) {
    return lang
  }

  const map = langMap.languages(extension)

  if (map?.length) {
    return MONACO_SUPPORTED_LANGUAGES.find(_lang => map.includes(_lang)) || PLAIN_TEXT
  }

  return PLAIN_TEXT
}

export type EnumPublicKeyUsage = 'auth'

export interface TypeKeys {
  created: number
  verified: number | undefined
  identifier: string
  usage: EnumPublicKeyUsage
  fingerprint: string
  comment: string
  type: string
}

interface WaitUtilParams<T> {
  test: () => T
  onMatched: (result: T) => void
  onExpired?: () => void
  duration?: number
  interval?: number
}

export interface inlineMergeFormValues {
  commitMessage: string
  commitTitle: string
}

export type inlineMergeFormRefType = FormikProps<inlineMergeFormValues>

export function waitUntil<T>({ test, onMatched, onExpired, duration = 5000, interval = 50 }: WaitUtilParams<T>) {
  const result = test()

  if (result) {
    onMatched(result)
  } else {
    if (duration > 0) {
      setTimeout(() => {
        waitUntil({ test, onMatched, onExpired, duration: duration - interval, interval })
      }, interval)
    } else {
      onExpired?.()
    }
  }
}

/**
 * Make a callback function that will be executed with no arg and return nothing.
 * @param f Any function.
 * @returns A callback function that accept no arguments and return nothing.
 */
// eslint-disable-next-line @typescript-eslint/ban-types
export const voidFn = (f: Function) => () => {
  f()
}

export enum MergeCheckStatus {
  MERGEABLE = 'mergeable',
  UNCHECKED = 'unchecked',
  CONFLICT = 'conflict'
}

export enum CodeOwnerReqDecision {
  CHANGEREQ = 'changereq',
  APPROVED = 'approved',
  WAIT_FOR_APPROVAL = ''
}
/**
 * Convert number of bytes into human readable format
 *
 * @param integer bytes     Number of bytes to convert
 * @param integer precision Number of digits after the decimal separator
 * @return string
 * @link https://stackoverflow.com/a/18650828/1114931
 */
export function formatBytes(bytes: number, decimals = 2) {
  if (!+bytes) return '0 Bytes'

  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['Bytes', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
}
export enum ScopeLevelEnum {
  ALL = 'all',
  CURRENT = 'current'
}

export enum PullRequestCheckType {
  EMPTY = '',
  RAW = 'raw',
  MARKDOWN = 'markdown',
  PIPELINE = 'pipeline'
}

/**
 * Test if an element is close to viewport. Used to determine a progressive
 * pre-rendering before the element being scrolled to viewport.
 */
export function isInViewport(element: Element, margin = 0, direction: 'x' | 'y' | 'xy' = 'y') {
  const rect = element.getBoundingClientRect()

  const height = window.innerHeight || document.documentElement.clientHeight
  const width = window.innerWidth || document.documentElement.clientWidth
  const _top = 0 - margin
  const _bottom = height + margin
  const _left = 0 - margin
  const _right = width + margin

  const yCheck =
    (rect.top >= _top && rect.top <= _bottom) ||
    (rect.bottom >= _top && rect.bottom <= _bottom) ||
    (rect.top <= _top && rect.bottom >= _bottom)
  const xCheck = (rect.left >= _left && rect.left <= _right) || (rect.right >= _top && rect.right <= _right)

  if (direction === 'y') return yCheck
  if (direction === 'x') return xCheck

  return yCheck || xCheck
}

export const truncateString = (str: string, length: number): string =>
  str.length <= length ? str : str.slice(0, length - 3) + '...'

export const isIterable = (value: unknown): boolean => {
  // checks for null and undefined
  if (value == null) {
    return false
  }
  return Symbol.iterator in Object(value)
}

/**
 *
 * @param obj - Object to get all keys for (including nested fields)
 * @param prefix - Optional prefix for keys
 * @returns all keys with their prefixes
 */
export const getAllKeysWithPrefix = (obj: { [key: string]: string | boolean | object }, prefix = ''): string[] => {
  let keys: string[] = []
  for (const key in obj) {
    if (Object.prototype.hasOwnProperty.call(obj, key)) {
      const currentKey = prefix ? `${prefix}.${key}` : key
      keys.push(currentKey)
      if (typeof obj[key] === 'object') {
        // Recursively get keys from nested objects with updated prefix
        keys = keys.concat(getAllKeysWithPrefix(obj[key] as { [key: string]: string | boolean | object }, currentKey))
      }
    }
  }
  return keys
}

export const CustomEventName = {
  SIDE_NAV_EXPANDED_EVENT: 'SIDE_NAV_EXPANDED_EVENT'
} as const

export const PAGE_CONTAINER_WIDTH = '--page-container-width'

// Outlets are used to insert additional components into CommentBox
export enum CommentBoxOutletPosition {
  START_OF_MARKDOWN_EDITOR_TOOLBAR = 'start_of_markdown_editor_toolbar',
  ENABLE_AIDA_PR_DESC_BANNER = 'enable_aida_pr_desc_banner'
}

// Helper function to escape special characters for use in regex
export function escapeRegExp(str: string) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') // $& means the whole matched string
}

export function removeSpecificTextOptimized(
  viewRef: React.MutableRefObject<EditorView | undefined>,
  textToRemove: string
) {
  const doc = viewRef?.current?.state.doc.toString()
  const regex = new RegExp(escapeRegExp(textToRemove), 'g')
  let match

  // Initialize an array to hold all changes
  const changes = []

  // Use regex to find all occurrences of the text
  while (doc && (match = regex.exec(doc)) !== null) {
    // Add a change object for each match to remove it
    changes.push({ from: match.index, to: match.index + match[0].length })
  }

  // Dispatch a single transaction with all changes if any matches were found
  if (changes.length > 0) {
    viewRef?.current?.dispatch({ changes })
  }
}

export const enum LabelType {
  DYNAMIC = 'dynamic',
  STATIC = 'static'
}

export enum ColorName {
  Red = 'red',
  Green = 'green',
  Yellow = 'yellow',
  Blue = 'blue',
  Pink = 'pink',
  Purple = 'purple',
  Violet = 'violet',
  Indigo = 'indigo',
  Cyan = 'cyan',
  Orange = 'orange',
  Brown = 'brown',
  Mint = 'mint',
  Lime = 'lime'
}

export interface ColorDetails {
  stroke: string
  background: string
  text: string
  backgroundWithoutStroke: string
}
export const colorsPanel: Record<ColorName, ColorDetails> = {
  [ColorName.Red]: { background: '#FFF7F7', stroke: '#F3A9AA', text: '#C7292F', backgroundWithoutStroke: '#FFE8EB' },
  [ColorName.Green]: { background: '#E9F9EE', stroke: '#85CBA2', text: '#16794C', backgroundWithoutStroke: '#E8F9ED' },
  [ColorName.Yellow]: { background: '#FFF9ED', stroke: '#E4B86F', text: '#92582D', backgroundWithoutStroke: '#FFF3D1' },
  [ColorName.Blue]: { background: '#F1FCFF', stroke: '#7BC8E0', text: '#236E93', backgroundWithoutStroke: '#E0F9FF' },
  [ColorName.Pink]: { background: '#FFF7FC', stroke: '#ECA8D2', text: '#C41B87', backgroundWithoutStroke: '#FEECF7' },
  [ColorName.Purple]: { background: '#FFF8FF', stroke: '#DFAAE3', text: '#9C2AAD', backgroundWithoutStroke: '#FCEDFC' },
  [ColorName.Violet]: { background: '#FBFAFF', stroke: '#C1B4F3', text: '#5645AF', backgroundWithoutStroke: '#F3F0FF' },
  [ColorName.Indigo]: { background: '#F8FAFF', stroke: '#A9BDF5', text: '#3250B2', backgroundWithoutStroke: '#EDF2FF' },
  [ColorName.Cyan]: { background: '#F2FCFD', stroke: '#7DCBD9', text: '#0B7792', backgroundWithoutStroke: '#E4F9FB' },
  [ColorName.Orange]: { background: '#FFF8F4', stroke: '#FFA778', text: '#995137', backgroundWithoutStroke: '#FFEDD5' },
  [ColorName.Brown]: { background: '#FCF9F6', stroke: '#DBB491', text: '#805C43', backgroundWithoutStroke: '#F8EFE7' },
  [ColorName.Mint]: { background: '#EFFEF9', stroke: '#7FD0BD', text: '#247469', backgroundWithoutStroke: '#DDFBF3' },
  [ColorName.Lime]: { background: '#F7FCF0', stroke: '#AFC978', text: '#586729', backgroundWithoutStroke: '#EDFADA' }
  // Add more colors when required
}

export const getColorsObj = (colorKey: ColorName): ColorDetails => {
  return colorsPanel[colorKey]
}
export const getScopeIcon = (scope: any, standalone: boolean) => {
  if (scope === 0) return undefined
  if (scope === 1 && standalone) return 'nav-project' as IconName
  if (scope === 1 && !standalone) return 'Account' as IconName
  if (scope === 2) return 'nav-organization' as IconName
  if (scope === 3) return 'nav-project' as IconName
  return undefined
}

export function customEncodeURIComponent(str: string) {
  return encodeURIComponent(str).replace(/!/g, '%21')
}

export interface LabelTypes extends TypesLabel {
  labelValues?: TypesLabelValue[]
}

export enum LabelFilterType {
  LABEL = 'label',
  VALUE = 'value'
}

export interface LabelFilterObj {
  labelId: number
  valueId?: number
  type: LabelFilterType
  labelObj: TypesLabel
  valueObj?: TypesLabelValue
}

export enum LabelsPageScope {
  ACCOUNT = 'acc',
  ORG = 'org',
  PROJECT = 'project',
  SPACE = 'space',
  REPOSITORY = 'repo'
}

export interface LabelListingProps {
  currentPageScope: LabelsPageScope
  repoMetadata?: RepoRepositoryOutput
  space?: string
  activeTab?: string
}

export const getScopeData = (space: string, scope: number, standalone: boolean) => {
  const accountIdentifier = space?.split('/')[0]
  const orgIdentifier = space?.split('/')[1]
  const projectIdentifier = space?.split('/')[2]
  if (standalone) {
    return { scopeRef: space, scopeIcon: 'nav-project' as IconName, scopeId: space }
  }
  switch (scope) {
    case 1:
      return { scopeRef: accountIdentifier as string, scopeIcon: 'Account' as IconName, scopeId: accountIdentifier }
    case 2:
      return {
        scopeRef: `${accountIdentifier}/${orgIdentifier}` as string,
        scopeIcon: 'nav-organization' as IconName,
        scopeId: orgIdentifier
      }
    case 3:
      return {
        scopeRef: `${accountIdentifier}/${orgIdentifier}/${projectIdentifier}` as string,
        scopeIcon: 'nav-project' as IconName,
        scopeId: projectIdentifier
      }
    default:
      return { scopeRef: space, scopeIcon: 'nav-project' as IconName, scopeId: scope }
  }
}

export enum RuleFields {
  APPROVALS_REQUIRE_MINIMUM_COUNT = 'pullreq.approvals.require_minimum_count',
  APPROVALS_REQUIRE_CODE_OWNERS = 'pullreq.approvals.require_code_owners',
  APPROVALS_REQUIRE_NO_CHANGE_REQUEST = 'pullreq.approvals.require_no_change_request',
  APPROVALS_REQUIRE_LATEST_COMMIT = 'pullreq.approvals.require_latest_commit',
  COMMENTS_REQUIRE_RESOLVE_ALL = 'pullreq.comments.require_resolve_all',
  STATUS_CHECKS_ALL_MUST_SUCCEED = 'pullreq.status_checks.all_must_succeed',
  STATUS_CHECKS_REQUIRE_IDENTIFIERS = 'pullreq.status_checks.require_identifiers',
  MERGE_STRATEGIES_ALLOWED = 'pullreq.merge.strategies_allowed',
  MERGE_DELETE_BRANCH = 'pullreq.merge.delete_branch',
  LIFECYCLE_CREATE_FORBIDDEN = 'lifecycle.create_forbidden',
  LIFECYCLE_DELETE_FORBIDDEN = 'lifecycle.delete_forbidden',
  MERGE_BLOCK = 'pullreq.merge.block',
  LIFECYCLE_UPDATE_FORBIDDEN = 'lifecycle.update_forbidden',
  LIFECYCLE_UPDATE_FORCE_FORBIDDEN = 'lifecycle.update_force_forbidden'
}
export type RuleFieldsMap = Record<RuleFields, boolean>

export type Rule = {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any
}

export type BranchProtectionRule = {
  title: string
  requiredRule: {
    [key in RuleFields]?: boolean
  }
}

export type BranchProtectionRulesMapType = Record<string, BranchProtectionRule>

export function createRuleFieldsMap(ruleDefinition: Rule): RuleFieldsMap {
  const ruleFieldsMap: RuleFieldsMap = {
    [RuleFields.APPROVALS_REQUIRE_MINIMUM_COUNT]: false,
    [RuleFields.APPROVALS_REQUIRE_CODE_OWNERS]: false,
    [RuleFields.APPROVALS_REQUIRE_NO_CHANGE_REQUEST]: false,
    [RuleFields.APPROVALS_REQUIRE_LATEST_COMMIT]: false,
    [RuleFields.COMMENTS_REQUIRE_RESOLVE_ALL]: false,
    [RuleFields.STATUS_CHECKS_ALL_MUST_SUCCEED]: false,
    [RuleFields.STATUS_CHECKS_REQUIRE_IDENTIFIERS]: false,
    [RuleFields.MERGE_STRATEGIES_ALLOWED]: false,
    [RuleFields.MERGE_DELETE_BRANCH]: false,
    [RuleFields.LIFECYCLE_CREATE_FORBIDDEN]: false,
    [RuleFields.LIFECYCLE_DELETE_FORBIDDEN]: false,
    [RuleFields.MERGE_BLOCK]: false,
    [RuleFields.LIFECYCLE_UPDATE_FORBIDDEN]: false,
    [RuleFields.LIFECYCLE_UPDATE_FORCE_FORBIDDEN]: false
  }
  if (ruleDefinition.pullreq) {
    if (ruleDefinition.pullreq.approvals) {
      ruleFieldsMap[RuleFields.APPROVALS_REQUIRE_CODE_OWNERS] = !!ruleDefinition.pullreq.approvals.require_code_owners
      ruleFieldsMap[RuleFields.APPROVALS_REQUIRE_LATEST_COMMIT] =
        !!ruleDefinition.pullreq.approvals.require_latest_commit
      ruleFieldsMap[RuleFields.APPROVALS_REQUIRE_MINIMUM_COUNT] =
        typeof ruleDefinition.pullreq.approvals.require_minimum_count === 'number'
      ruleFieldsMap[RuleFields.APPROVALS_REQUIRE_NO_CHANGE_REQUEST] =
        !!ruleDefinition.pullreq.approvals.require_no_change_request
    }

    if (ruleDefinition.pullreq.comments) {
      ruleFieldsMap[RuleFields.COMMENTS_REQUIRE_RESOLVE_ALL] = !!ruleDefinition.pullreq.comments.require_resolve_all
    }

    if (ruleDefinition.pullreq.merge) {
      ruleFieldsMap[RuleFields.MERGE_BLOCK] = !!ruleDefinition.pullreq.merge.block
      ruleFieldsMap[RuleFields.MERGE_DELETE_BRANCH] = !!ruleDefinition.pullreq.merge.delete_branch
      ruleFieldsMap[RuleFields.MERGE_STRATEGIES_ALLOWED] =
        Array.isArray(ruleDefinition.pullreq.merge.strategies_allowed) &&
        ruleDefinition.pullreq.merge.strategies_allowed.length > 0
    }

    if (ruleDefinition.pullreq.status_checks) {
      ruleFieldsMap[RuleFields.STATUS_CHECKS_REQUIRE_IDENTIFIERS] =
        Array.isArray(ruleDefinition.pullreq.status_checks.require_identifiers) &&
        ruleDefinition.pullreq.status_checks.require_identifiers.length > 0
    }
  }

  if (ruleDefinition.lifecycle) {
    ruleFieldsMap[RuleFields.LIFECYCLE_CREATE_FORBIDDEN] = !!ruleDefinition.lifecycle.create_forbidden
    ruleFieldsMap[RuleFields.LIFECYCLE_DELETE_FORBIDDEN] = !!ruleDefinition.lifecycle.delete_forbidden
    ruleFieldsMap[RuleFields.LIFECYCLE_UPDATE_FORBIDDEN] = !!ruleDefinition.lifecycle.update_forbidden
    ruleFieldsMap[RuleFields.LIFECYCLE_UPDATE_FORCE_FORBIDDEN] = !!ruleDefinition.lifecycle.update_force_forbidden
  }

  return ruleFieldsMap
}
