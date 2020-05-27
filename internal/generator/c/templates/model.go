/*
 * Copyright 2019 ObjectBox Ltd. All rights reserved.
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

package templates

import (
	"text/template"
)

// ModelTemplate is used to generate the model initialization code
// TODO property relations, indexes
var ModelTemplate = template.Must(template.New("model").Funcs(funcMap).Parse(
	`// Code generated by ObjectBox; DO NOT EDIT.

#ifndef OBJECTBOX_MODEL_H
#define OBJECTBOX_MODEL_H

#include <stdbool.h>
#include <stdint.h>

#include "objectbox.h"

#ifdef __cplusplus
extern "C" {
#endif

inline OBX_model* create_obx_model() {
    OBX_model* model = obx_model();
    if (!model) return NULL;

	bool successful = false;
	do {
		{{- range $entity := .Model.Entities}}
		if (obx_model_entity(model, "{{$entity.Name}}", {{$entity.Id.GetId}}, {{$entity.Id.GetUid}})) break;
		{{range $property := $entity.Properties -}}
		if (obx_model_property(model, "{{$property.Name}}", {{CorePropType $property.Type}}, {{$property.Id.GetId}}, {{$property.Id.GetUid}})) break;
		{{with $property.Flags}}if (obx_model_property_flags(model, {{CorePropFlags .}})) break;
		{{end -}}
		{{end -}}
		if (obx_model_entity_last_property_id(model, {{$entity.LastPropertyId.GetId}}, {{$entity.LastPropertyId.GetUid}})) break;
		{{end -}}

		obx_model_last_entity_id(model, {{.Model.LastEntityId.GetId}}, {{.Model.LastEntityId.GetUid}});
		{{if .Model.LastIndexId}}obx_model_last_index_id(model, {{.Model.LastIndexId.GetId}}, {{.Model.LastIndexId.GetUid}});{{end -}}
		{{if .Model.LastRelationId}}obx_model_last_relation_id(model, {{.Model.LastRelationId.GetId}}, {{.Model.LastRelationId.GetUid}});{{end -}}

		successful = true;
	} while (false);

	if (!successful) {
		// TODO error handling 
		// obx_model_error_message(model);
		obx_model_free(model);
        return NULL;
	}

	return model;
}

#ifdef __cplusplus
}
#endif

#endif  // OBJECTBOX_MODEL_H
`))
