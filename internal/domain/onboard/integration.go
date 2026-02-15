package onboard

import (
	"errors"
	"slices"
)

type IntegrationComponent string

func ParseIntegrationComponent(policy *IntegrationPolicy, component string) (IntegrationComponent, error) {
	if slices.Contains(policy.components, IntegrationComponent(component)) {
		return IntegrationComponent(component), nil
	}

	return "", ErrInvalidIntegrationComponent
}

type Integration struct {
	onboardID  OnboardID
	components map[IntegrationComponent]bool
}

func newIntegration(policy *IntegrationPolicy, onboardID OnboardID) Integration {

	components := make(map[IntegrationComponent]bool)
	for _, name := range policy.components {
		components[IntegrationComponent(name)] = false
	}

	return Integration{
		onboardID:  onboardID,
		components: components,
	}
}

func (i *Integration) MarkComponentReady(policy *IntegrationPolicy, component IntegrationComponent) error {
	if isReady := i.components[component]; isReady {
		return ErrIntegrationComponentAlreadyReady
	}

	i.components[component] = true
	return nil
}

func (i *Integration) IsCompleted(policy *IntegrationPolicy) bool {
	for _, requiredComponent := range policy.components {
		if !i.components[requiredComponent] {
			return false
		}
	}

	return true
}

func RebuildIntegration(onboardID OnboardID, components map[string]bool) (Integration, error) {
	integration := Integration{
		onboardID:  onboardID,
		components: make(map[IntegrationComponent]bool),
	}

	var errs error

	for component, isReady := range components {
		if component == " " {
			errs = errors.Join(errs, errors.New("invalid (blanck) component"))
		} else {
			integration.components[IntegrationComponent(component)] = isReady
		}
	}

	if errs != nil {
		return Integration{}, errs
	}

	return integration, nil
}
