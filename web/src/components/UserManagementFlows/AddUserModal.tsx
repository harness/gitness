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

import React, { useState } from 'react'
import {
  Button,
  ButtonVariation,
  Container,
  Dialog,
  FlexExpander,
  FormikForm,
  FormInput,
  Layout,
  Text,
  useToaster
} from '@harnessio/uicore'
import { Icon } from '@harnessio/icons'
import { Color, FontVariation } from '@harnessio/design-system'
import { Formik, FormikHelpers, useFormikContext } from 'formik'
import { get } from 'lodash-es'
import { Else, Match, Render, Truthy } from 'react-jsx-match'
import * as Yup from 'yup'
import { useModalHook } from 'hooks/useModalHook'
import { useStrings } from 'framework/strings'
import { OpenapiAdminUsersCreateRequest, TypesUser, useAdminCreateUser, useAdminUpdateUser } from 'services/code'
import { generateAlphaNumericHash, getErrorMessage, REGEX_VALID_USER_ID } from 'utils/Utils'
import { CodeIcon } from 'utils/GitUtils'
import { CopyButton } from 'components/CopyButton/CopyButton'

import css from './UserManagementFlows.module.scss'

const useAddUserModal = ({ onClose }: { onClose: () => Promise<void> }) => {
  const { getString } = useStrings()
  const { showSuccess, showError } = useToaster()

  const [isEditMode, setIsEditMode] = useState(false)
  const [userDetails, setUserDetails] = useState<TypesUser>()
  const [isUserCreated, setIsUserCreated] = useState(false)

  const { mutate: createUser } = useAdminCreateUser({})
  const { mutate: updateUser } = useAdminUpdateUser({
    user_uid: userDetails?.uid || ''
  })

  const onSubmit = async (values: TypesUser, formikHelpers: FormikHelpers<TypesUser>) => {
    try {
      const mutate = isEditMode ? updateUser : createUser
      let payload: OpenapiAdminUsersCreateRequest = {
        ...values
      }

      if (!payload.display_name) {
        payload.display_name = payload.uid
        formikHelpers.setFieldValue('display_name', payload.uid)
      }

      if (!isEditMode) {
        const password = generateAlphaNumericHash(10)
        formikHelpers.setFieldValue('password', password)

        payload = {
          ...payload,
          password
        }

        setIsUserCreated(true)
      }

      await mutate(payload)

      if (isEditMode) {
        showSuccess(getString('newUserModal.userUpdated'))
        hideModal()
        onClose()
      } else {
        showSuccess(getString('newUserModal.userCreated'))
      }
    } catch (error) {
      showError(getErrorMessage(error))
    }
  }

  const onModalClose = () => {
    hideModal()
    onClose()
    setIsUserCreated(false)
  }

  const [openModal, hideModal] = useModalHook(
    () => (
      <Dialog
        isOpen
        enforceFocus={false}
        onClose={onModalClose}
        style={{ height: isEditMode ? 420 : 530 }}
        title={isEditMode ? getString('updateUser') : getString('createNewUser')}
        chidrenClassName={css.dialogCtn}>
        <Formik<OpenapiAdminUsersCreateRequest>
          validationSchema={Yup.object().shape({
            uid: Yup.string()
              .required(getString('validation.uidRequired'))
              .matches(REGEX_VALID_USER_ID, getString('validation.uidInvalid')),
            email: Yup.string()
              .required(getString('validation.emailRequired'))
              .email(getString('validation.emailInvalid')),
            display_name: Yup.string()
          })}
          initialValues={{
            uid: userDetails?.uid || '',
            display_name: userDetails?.display_name || '',
            email: userDetails?.email || ''
          }}
          onSubmit={onSubmit}>
          {formikProps => {
            return (
              <FormikForm className={css.formikForm}>
                <FormInputWithCopyButton
                  name="uid"
                  placeholder={getString('newUserModal.uidPlaceholder')}
                  label={
                    <Layout.Horizontal style={{ alignItems: 'center' }} spacing="small">
                      <Text color={Color.GREY_600} font={{ variation: FontVariation.BODY2 }}>
                        {getString('userId')}
                      </Text>
                      <Render when={!isEditMode}>
                        <Text
                          icon="tooltip-icon"
                          iconProps={{ size: 12, color: Color.PRIMARY_7 }}
                          font={{ variation: FontVariation.BODY2 }}
                          color={Color.PRIMARY_7}>
                          {getString('newUserModal.uidWarning')}
                        </Text>
                      </Render>
                    </Layout.Horizontal>
                  }
                  disabled={isEditMode || isUserCreated}
                />
                <FormInputWithCopyButton
                  name="email"
                  label={getString('email')}
                  placeholder={getString('newUserModal.emailPlaceholder')}
                  disabled={isUserCreated}
                />
                <FormInputWithCopyButton
                  name="display_name"
                  label={getString('displayName')}
                  placeholder={getString('newUserModal.displayNamePlaceholder')}
                  disabled={isUserCreated}
                  isOptional
                />
                <Render when={isUserCreated}>
                  <GeneratedPassword password={formikProps.values.password || ''} />
                </Render>
                <FlexExpander />
                <Container>
                  <Match expr={isUserCreated}>
                    <Truthy>
                      <Button text={getString('close')} variation={ButtonVariation.TERTIARY} onClick={onModalClose} />
                    </Truthy>
                    <Else>
                      <Button
                        type="submit"
                        text={isEditMode ? getString('save') : getString('createUser')}
                        variation={ButtonVariation.PRIMARY}
                      />
                    </Else>
                  </Match>
                </Container>
              </FormikForm>
            )
          }}
        </Formik>
      </Dialog>
    ),
    [isEditMode, onSubmit, userDetails, isUserCreated]
  )

  return {
    openModal: ({ isEditing, userInfo }: { isEditing?: boolean; userInfo?: TypesUser }) => {
      openModal()
      setIsEditMode(Boolean(isEditing))
      setUserDetails(userInfo)
    }
  }
}

export default useAddUserModal

export const GeneratedPassword: React.FC<{ password: string }> = ({ password }) => {
  const { getString } = useStrings()

  return (
    <Container background={Color.GREEN_50} className={css.passwordCtn}>
      <Text color={Color.GREY_600} font={{ variation: FontVariation.BODY2 }} margin={{ bottom: 'xsmall' }}>
        {getString('password')}
      </Text>
      <Layout.Horizontal style={{ alignItems: 'center' }} margin={{ bottom: 'small' }}>
        <Layout.Horizontal className={css.layout}>
          <Text className={css.text}>{password}</Text>
          <FlexExpander />
          <CopyButton content={password} id={css.copyBtn} icon={CodeIcon.Copy} iconProps={{ size: 14 }} />
        </Layout.Horizontal>
        <Icon name="command-artifact-check" color={Color.GREEN_700} padding={{ left: 'small' }} />
      </Layout.Horizontal>
      <Text font={{ variation: FontVariation.TINY_SEMI }} color={Color.GREEN_900}>
        {getString('newUserModal.passwordHelptext')}
      </Text>
    </Container>
  )
}

export const FormInputWithCopyButton: React.FC<
  React.ComponentProps<typeof FormInput.Text> & { copyContent?: string }
> = props => {
  const { values } = useFormikContext() || {}

  return (
    <FormInput.Text
      {...props}
      className={css.inputWrapper}
      inputGroup={{
        ...props.inputGroup,
        rightElement: (
          <Render when={props.disabled}>
            <CopyButton
              content={get(values, props.name, props.copyContent || '')}
              id={css.copyBtn}
              icon={CodeIcon.Copy}
              iconProps={{ size: 14 }}
            />
          </Render>
        )
      }}
    />
  )
}
