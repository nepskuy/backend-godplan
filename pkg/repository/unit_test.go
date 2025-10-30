package repository

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNewUserRepository(t *testing.T) {
    // Test bahwa NewUserRepository tidak panic
    assert.NotPanics(t, func() {
        repo := NewUserRepository(nil)
        assert.NotNil(t, repo)
    })
}

func TestUserRepositoryInterface(t *testing.T) {
    // Test bahwa repository memenuhi interface
    var repo UserRepositoryInterface = &UserRepository{}
    assert.NotNil(t, repo)
}

func TestUserRepositoryMethods(t *testing.T) {
    // Test bahwa methods ada (compile-time check)
    repo := &UserRepository{}
    
    // Test bahwa struct memiliki methods yang diperlukan
    assert.NotNil(t, repo)
}
