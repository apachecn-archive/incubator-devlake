/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import React from 'react'
import {
  Button,
  Icon,
  Intent,
  Switch,
  FormGroup,
  ButtonGroup,
  RadioGroup,
  Radio,
  InputGroup,
  Divider,
  Elevation,
  TextArea,
  Tabs,
  Tab,
  Card,
  Popover,
  Tooltip,
  Label,
  MenuItem,
  Position,
  Colors,
  Tag,
} from '@blueprintjs/core'
import { MultiSelect, Select } from '@blueprintjs/select'
import InputValidationError from '@/components/validation/InputValidationError'

const DataEntitiesSelector = (props) => {
  const {
    connections = [],
    configuredConnection,
    placeholder = 'Select data entities',
    items = [],
    selectedItems = [],
    restrictedItems = [],
    activeItem = null,
    disabled = false,
    isSaving = false,
    onItemSelect = () => {},
    onRemove = () => {},
    onClear = () => {},
    fieldHasError = () => {},
    getFieldError = () => {},
    itemRenderer = (item, { handleClick, modifiers }) => (
      <MenuItem
        active={modifiers.active || selectedItems.find(i => i.id === item.id)}
        disabled={
          selectedItems.find(i => i.id === item.id)
        }
        key={item.value}
        // label=
        onClick={handleClick}
        text={
          selectedItems.find(i => i.id === item.id)
            ? (
              <>
                <input type='checkbox' checked readOnly /> {item.title}
              </>
              )
            : (
              <span style={{ fontWeight: 700 }}>
                <input type='checkbox' readOnly /> {item.title}
              </span>
              )
        }
        style={{
          marginBottom: '2px',
          fontWeight: items.includes(item) ? 700 : 'normal',
        }}
      />
    ),
    tagRenderer = (item) => item.title,
  } = props
  return (
    <>
      <div
        className='data-entities-multiselect'
        style={{ display: 'flex', marginBottom: '10px' }}
      >
        <div
          className='data-entities-multiselect-selector'
          style={{ minWidth: '200px', width: '100%' }}
        >
          <MultiSelect
            disabled={disabled || isSaving}
            // openOnKeyDown={true}
            resetOnSelect={true}
            placeholder={placeholder}
            popoverProps={{ usePortal: false, minimal: true }}
            className='multiselector-entities'
            inline={true}
            fill={true}
            items={items}
            selectedItems={selectedItems}
            activeItem={activeItem}
            itemPredicate={(query, item) =>
              item?.title.toLowerCase().indexOf(query.toLowerCase()) >= 0}
            itemRenderer={itemRenderer}
            tagRenderer={tagRenderer}
            tagInputProps={{
              tagProps: {
                intent: Intent.PRIMARY,
                minimal: true,
              },
            }}
            noResults={<MenuItem disabled={true} text='No Data Entities.' />}
            onRemove={(item) => {
              onRemove((rT) => {
                return {
                  ...rT,
                  [configuredConnection.id]: rT[configuredConnection.id].filter(
                    (t) => t.id !== item.id
                  ),
                }
              })
            }}
            onItemSelect={(item) => {
              onItemSelect((rT) => {
                return !rT[configuredConnection.id].includes(item)
                  ? {
                      ...rT,
                      [configuredConnection.id]: [
                        ...rT[configuredConnection.id],
                        item,
                      ],
                    }
                  : { ...rT }
              })
            }}
            style={{ borderRight: 0 }}
          />
        </div>
      </div>
    </>
  )
}

export default DataEntitiesSelector
