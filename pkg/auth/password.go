// percona-everest-backend
// Copyright (C) 2023 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package auth holds logic for authentication.
package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/pbkdf2"
	corev1 "k8s.io/api/core/v1"
)

// Password supports authentication by providing a password
// and comparing it to a hash stored in Kubernetes.
type Password struct {
	kubeClient kubeClient
	l          *zap.SugaredLogger

	// Guards hash and refreshedAt
	mu          sync.RWMutex
	hash        string
	refreshedAt time.Time

	namespaceUID []byte
}

type kubeClient interface {
	GetSecret(ctx context.Context, name string) (*corev1.Secret, error)
}

const hashExpiration = 3 * time.Second

// NewPassword returns a new Password struct.
func NewPassword(k kubeClient, l *zap.SugaredLogger, namespaceUID []byte) *Password {
	return &Password{
		kubeClient:   k,
		l:            l,
		namespaceUID: namespaceUID,
	}
}

// Valid returns true if the provided password is valid/correct.
func (p *Password) Valid(ctx context.Context, password string) (bool, error) {
	if password == "" {
		return false, nil
	}

	storedHash, err := p.hashFromSecret(ctx)
	if err != nil {
		return false, errors.Join(err, errors.New("could not validate password against the stored hash"))
	}

	salt := p.namespaceUID
	hash := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)

	if string(hash) == storedHash {
		return true, nil
	}

	return false, nil
}

func (p *Password) hashFromSecret(ctx context.Context) (string, error) {
	p.mu.RLock()

	if !p.refreshedAt.IsZero() && time.Now().Before(p.refreshedAt.Add(hashExpiration)) {
		defer p.mu.RUnlock()
		p.l.Debug("Using cached password hash")

		return p.hash, nil
	}

	p.mu.RUnlock()
	p.mu.Lock()
	defer p.mu.Unlock()

	hash, err := p.hashFromK8s(ctx)
	if err != nil {
		return "", err
	}

	p.hash = string(hash)
	p.refreshedAt = time.Now()

	return p.hash, nil
}

func (p *Password) hashFromK8s(ctx context.Context) ([]byte, error) {
	p.l.Debug("Getting password hash from k8s")

	secret, err := p.kubeClient.GetSecret(ctx, "everest-password")
	if err != nil {
		return nil, errors.Join(err, errors.New("could not get stored password from Kubernetes"))
	}
	storedHash, ok := secret.Data["password"]
	if !ok {
		return nil, errors.Join(err, errors.New("could not get stored password hash from secret"))
	}

	return storedHash, nil
}
