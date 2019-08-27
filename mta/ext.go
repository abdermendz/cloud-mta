package mta

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	unmarshalExtErrorMsg   = "could not unmarshal the MTA extension file"
	mergeModulesErrorMsg   = `could not merge modules from MTA extension with ID "%s"`
	mergeResourcesErrorMsg = `could not merge resources from MTA extension with ID "%s"`

	mergeRootParametersErrorMsg = `could not merge parameters from MTA extension with ID "%s"`

	mergeModulePropertiesErrorMsg             = `could not merge the properties of the "%s" module`
	mergeModuleParametersErrorMsg             = `could not merge the parameters of the "%s" module`
	mergeModuleBuildParametersErrorMsg        = `could not merge the build parameters of the "%s" module`
	mergeModuleIncludesErrorMsg               = `could not merge the includes of the "%s" module`
	mergeModuleProvidesPropertiesErrorMsg     = `could not merge the properties of the "%s" provides in the "%s" module`
	mergeModuleRequiresPropertiesErrorMsg     = `could not merge the properties of the "%s" requires in the "%s" module`
	mergeModuleRequiresParametersErrorMsg     = `could not merge the parameters of the "%s" requires in the "%s" module`
	mergeModuleHookParametersErrorMsg         = `could not merge the parameters of the "%s" hook in the "%s" module`
	mergeModuleHookRequiresPropertiesErrorMsg = `could not merge the properties of the "%s" requires in the "%s" hook of the "%s" module`
	mergeModuleHookRequiresParametersErrorMsg = `could not merge the parameters of the "%s" requires in the "%s" hook of the "%s" module`
	unknownModuleHookRequiresErrorMsg         = `the "%s" requires in the "%s" hook of the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`
	unknownModuleProvidesErrorMsg             = `the "%s" provides in the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`
	unknownModuleRequiresErrorMsg             = `the "%s" requires in the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`
	unknownModuleHookErrorMsg                 = `the "%s" hook in the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`
	unknownModuleErrorMsg                     = `the "%s" module is defined in the MTA extension but not in the "mta.yaml" file`

	mergeResourceActiveErrorMsg             = `could not merge the active property of the "%s" resource`
	mergeResourcePropertiesErrorMsg         = `could not merge the properties of the "%s" resource`
	mergeResourceParametersErrorMsg         = `could not merge the parameters of the "%s" resource`
	mergeResourceRequiresPropertiesErrorMsg = `could not merge the properties of the "%s" requires in the "%s" resource`
	mergeResourceRequiresParametersErrorMsg = `could not merge the parameters of the "%s" requires in the "%s" resource`
	unknownResourceRequiresErrorMsg         = `the "%s" requires in the "%s" resource is defined in the MTA extension but not in the "mta.yaml" file`
	unknownResourceErrorMsg                 = `the "%s" resource is defined in the MTA extension but not in the "mta.yaml" file`

	overwriteStructuredWithScalarErrorMsg = `"%s": cannot overwrite a structured value with a scalar value`
	overwriteScalarWithStructuredErrorMsg = `"%s": cannot overwrite a scalar value with a structured value`
	overwriteNonOverwritableErrorMsg      = `the "%s" field is not overwritable`
)

// UnmarshalExt returns a reference to the EXT object from a byte array.
func UnmarshalExt(content []byte) (*EXT, error) {
	var m EXT
	// Unmarshal MTA file
	err := yaml.Unmarshal(content, &m)
	if err != nil {
		err = errors.Wrap(err, unmarshalExtErrorMsg)
	}
	return &m, err
}

// Merge merges mta object with mta extension object extension properties complement and overwrite mta properties
func Merge(mta *MTA, mtaExt *EXT) error {
	err := chain().
		extendMap(&mta.Parameters, mta.ParametersMetaData, mtaExt.Parameters, mergeRootParametersErrorMsg, mtaExt.ID).
		err
	if err != nil {
		return err
	}

	if err = mergeModules(*mta, mtaExt.Modules); err != nil {
		return errors.Wrapf(err, mergeModulesErrorMsg, mtaExt.ID)
	}

	if err = mergeResources(*mta, mtaExt.Resources); err != nil {
		return errors.Wrapf(err, mergeResourcesErrorMsg, mtaExt.ID)
	}

	return nil
}

// mergeModules is responsible for handling the rules of merging modules
func mergeModules(mtaObj MTA, mtaExtModules []*ModuleExt) error {
	for _, extModule := range mtaExtModules {
		if module, err := mtaObj.GetModuleByName(extModule.Name); err == nil {
			err = chain().
				extendMap(&module.Properties, module.PropertiesMetaData, extModule.Properties, mergeModulePropertiesErrorMsg, module.Name).
				extendMap(&module.Parameters, module.ParametersMetaData, extModule.Parameters, mergeModuleParametersErrorMsg, module.Name).
				extendMap(&module.BuildParams, nil, extModule.BuildParams, mergeModuleBuildParametersErrorMsg, module.Name).
				extendIncludes(&module.Includes, extModule.Includes, mergeModuleIncludesErrorMsg, module.Name).
				err
			if err != nil {
				return err
			}
			err = mergeModuleProvides(module, extModule)
			if err != nil {
				return err
			}
			if err = mergeRequires(extModule.Requires, module,
				msg{unknownModuleRequiresErrorMsg, []interface{}{extModule.Name}},
				msg{mergeModuleRequiresPropertiesErrorMsg, []interface{}{module.Name}},
				msg{mergeModuleRequiresParametersErrorMsg, []interface{}{module.Name}}); err != nil {
				return err
			}
			err = mergeModuleHooks(module, extModule)
			if err != nil {
				return err
			}
		} else {
			return errors.Wrapf(err, unknownModuleErrorMsg, extModule.Name)
		}
	}

	return nil
}

func mergeModuleProvides(module *Module, extModule *ModuleExt) error {
	for _, extProvide := range extModule.Provides {
		if provide := module.GetProvidesByName(extProvide.Name); provide != nil {
			err := chain().
				extendMap(&provide.Properties, provide.PropertiesMetaData, extProvide.Properties, mergeModuleProvidesPropertiesErrorMsg, provide.Name, module.Name).
				err
			if err != nil {
				return err
			}
		} else {
			return errors.Errorf(unknownModuleProvidesErrorMsg, extProvide.Name, extModule.Name)
		}
	}
	return nil
}

func mergeModuleHooks(module *Module, extModule *ModuleExt) error {
	for _, extHook := range extModule.Hooks {
		if hook := module.GetHookByName(extHook.Name); hook != nil {
			err := chain().
				extendMap(&hook.Parameters, hook.ParametersMetaData, extHook.Parameters, mergeModuleHookParametersErrorMsg, hook.Name, module.Name).
				err
			if err != nil {
				return err
			}
			if err = mergeRequires(extHook.Requires, hook,
				msg{unknownModuleHookRequiresErrorMsg, []interface{}{extHook.Name, extModule.Name}},
				msg{mergeModuleHookRequiresPropertiesErrorMsg, []interface{}{hook.Name, module.Name}},
				msg{mergeModuleHookRequiresParametersErrorMsg, []interface{}{hook.Name, module.Name}}); err != nil {
				return err
			}
		} else {
			return errors.Errorf(unknownModuleHookErrorMsg, extHook.Name, extModule.Name)
		}
	}
	return nil
}

// mergeResources is responsible for handling the rules of merging resources
func mergeResources(mtaObj MTA, mtaExtResources []*ResourceExt) error {
	for _, extResource := range mtaExtResources {
		if resource, err := mtaObj.GetResourceByName(extResource.Name); err == nil {
			err = chain().
				extendBool(&resource.Active, &extResource.Active, mergeResourceActiveErrorMsg, resource.Name).
				extendMap(&resource.Properties, resource.PropertiesMetaData, extResource.Properties, mergeResourcePropertiesErrorMsg, resource.Name).
				extendMap(&resource.Parameters, resource.ParametersMetaData, extResource.Parameters, mergeResourceParametersErrorMsg, resource.Name).
				err
			if err != nil {
				return err
			}
			if err = mergeRequires(extResource.Requires, resource,
				msg{unknownResourceRequiresErrorMsg, []interface{}{extResource.Name}},
				msg{mergeResourceRequiresPropertiesErrorMsg, []interface{}{resource.Name}},
				msg{mergeResourceRequiresParametersErrorMsg, []interface{}{resource.Name}}); err != nil {
				return err
			}
		} else {
			return errors.Wrapf(err, unknownResourceErrorMsg, extResource.Name)
		}
	}

	return nil
}

type requiresProvider interface {
	GetRequiresByName(name string) *Requires
}
type msg struct {
	msg  string
	args []interface{}
}

func (msg msg) getArgs(prependArgs ...interface{}) []interface{} {
	return append(prependArgs, msg.args...)
}

// mergeRequires is responsible for merging the requires part of modules, resources etc
func mergeRequires(requires []Requires, extRequiresProvider requiresProvider, unknownRequiresMsg msg, mergePropertiesMsg msg, mergeParametersMsg msg) error {
	for _, extRequires := range requires {
		if requires := extRequiresProvider.GetRequiresByName(extRequires.Name); requires != nil {
			err := chain().
				extendMap(&requires.Properties, requires.PropertiesMetaData, extRequires.Properties, mergePropertiesMsg.msg, mergePropertiesMsg.getArgs(requires.Name)...).
				extendMap(&requires.Parameters, requires.ParametersMetaData, extRequires.Parameters, mergeParametersMsg.msg, mergeParametersMsg.getArgs(requires.Name)...).
				err
			if err != nil {
				return err
			}
		} else {
			return errors.Errorf(unknownRequiresMsg.msg, unknownRequiresMsg.getArgs(extRequires.Name)...)
		}
	}
	return nil
}

// isFieldOverWritable test is the current field allowed to be overwritten in mta file
func isFieldOverWritable(field string, meta map[string]MetaData, m map[string]interface{}) bool {
	if meta != nil && m[field] != nil {
		if metaData, exists := meta[field]; exists {
			return metaData.OverWritable
		}
	}
	return true
}

// extendMap extends map with elements of mta extension map
func extendMap(m *map[string]interface{}, meta map[string]MetaData, ext map[string]interface{}) error {
	if ext != nil {
		if *m == nil {
			*m = make(map[string]interface{})
		}
		for key, value := range ext {
			if isFieldOverWritable(key, meta, *m) {
				err := mergeMapKey(m, key, value)
				if err != nil {
					return err
				}
			} else {
				return errors.Errorf(overwriteNonOverwritableErrorMsg, key)
			}
		}
	}
	return nil
}

func mergeMapKey(m *map[string]interface{}, key string, value interface{}) error {
	extMapValue, extIsMap := value.(map[string]interface{})
	mMapValue, mIsMap := (*m)[key].(map[string]interface{})
	if (*m)[key] == nil || value == nil || (!extIsMap && !mIsMap) {
		(*m)[key] = value
	} else if mIsMap && extIsMap {
		if err := extendMap(&mMapValue, nil, extMapValue); err != nil {
			return errors.Wrapf(err, `"%s"`, key)
		}
	} else {
		// Both values aren't nil. One of them is a map and the other is not.
		if mIsMap {
			return errors.Errorf(overwriteStructuredWithScalarErrorMsg, key)
		}
		return errors.Errorf(overwriteScalarWithStructuredErrorMsg, key)
	}
	return nil
}

// extendIncludes extends Slice Includes with elements of mta extension Slice Includes
func extendIncludes(m *[]Includes, ext []Includes) {
	if ext != nil {
		if *m == nil {
			*m = []Includes{}
		}
		*m = append(*m, ext...)
	}
}

// extendBool extends bool with element of mta extension bool
func extendBool(m *bool, ext *bool) {
	if ext != nil {
		*m = *ext
	}
}

func chain() *chainError {
	return &chainError{}
}

// chainError is a struct for chaining error from the different function, allowing them to be called in one line.
type chainError struct {
	err error
}

func (v *chainError) extendMap(m *map[string]interface{}, meta map[string]MetaData, ext map[string]interface{}, msg string, args ...interface{}) *chainError {
	if v.err != nil {
		return v
	}
	err := extendMap(m, meta, ext)
	if err != nil {
		v.err = errors.Wrapf(err, msg, args...)
	}
	return v
}

func (v *chainError) extendIncludes(m *[]Includes, ext []Includes, msg string, args ...interface{}) *chainError {
	if v.err != nil {
		return v
	}
	extendIncludes(m, ext)
	return v
}

func (v *chainError) extendBool(m *bool, ext *bool, msg string, args ...interface{}) *chainError {
	if v.err != nil {
		return v
	}
	extendBool(m, ext)
	return v
}
