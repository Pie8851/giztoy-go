//go:build gizclaw_e2e

package rpc_test

import (
	"testing"

	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/rpcapi"
)

func TestServerPetRPC(t *testing.T) {
	assertRemovedBusinessRPCSurfaces(t)
	env := newBusinessHarness(t)

	petID := "pet-a"
	pet, err := env.a.AdoptPet(env.ctx, "pet.adopt", rpcapi.PetAdoptRequest{Id: &petID, Name: "Momo"})
	if err != nil {
		t.Fatalf("pet.adopt: %v", err)
	}
	if pet.Id != "pet-a" || pet.SpeciesId != "rabbit" || pet.VoiceId != "voice-a" {
		t.Fatalf("adopted pet = %#v", pet)
	}
	pet, err = env.a.FeedPet(env.ctx, "pet.feed", rpcapi.PetFeedRequest{PetId: "pet-a", Prompt: "ate lunch"})
	if err != nil {
		t.Fatalf("pet.feed: %v", err)
	}
	if pet.Life.Satiety < 60 {
		t.Fatalf("pet.feed satiety = %d, want increased", pet.Life.Satiety)
	}
	pet, err = env.a.WashPet(env.ctx, "pet.wash", rpcapi.PetWashRequest{PetId: "pet-a", Prompt: "had a bath"})
	if err != nil {
		t.Fatalf("pet.wash: %v", err)
	}
	if pet.Life.Cleanliness < 60 {
		t.Fatalf("pet.wash cleanliness = %d, want increased", pet.Life.Cleanliness)
	}
	pet, err = env.a.PlayPet(env.ctx, "pet.play", rpcapi.PetPlayRequest{PetId: "pet-a", Prompt: "played a game"})
	if err != nil {
		t.Fatalf("pet.play: %v", err)
	}
	if pet.Life.Mood < 57 || pet.Ability.Exp == 0 {
		t.Fatalf("pet.play result = %#v", pet)
	}

	renamedPet, err := env.a.PutPet(env.ctx, "pet.put", rpcapi.PetPutRequest{Id: "pet-a", Name: "Momo II"})
	if err != nil {
		t.Fatalf("pet.put: %v", err)
	}
	if renamedPet.Name != "Momo II" {
		t.Fatalf("pet.put name = %q, want Momo II", renamedPet.Name)
	}
	gotPet, err := env.a.GetPet(env.ctx, "pet.get", rpcapi.PetGetRequest{Id: "pet-a"})
	if err != nil {
		t.Fatalf("pet.get: %v", err)
	}
	if gotPet.Id != "pet-a" || gotPet.Name != "Momo II" {
		t.Fatalf("pet.get = %#v", gotPet)
	}
	secondPetID := "pet-b"
	secondPet, err := env.a.AdoptPet(env.ctx, "pet.adopt.second", rpcapi.PetAdoptRequest{Id: &secondPetID, Name: "Nono"})
	if err != nil {
		t.Fatalf("pet.adopt second: %v", err)
	}
	assertPetPagination(t, env.ctx, env.a, []string{"pet-a", secondPet.Id})
	if _, err := env.b.GetPet(env.ctx, "pet.get.denied", rpcapi.PetGetRequest{Id: "pet-a"}); err == nil {
		t.Fatalf("pet.get from peer-b should fail")
	}
	deletedPet, err := env.a.DeletePet(env.ctx, "pet.delete", rpcapi.PetDeleteRequest{Id: secondPet.Id})
	if err != nil {
		t.Fatalf("pet.delete: %v", err)
	}
	if deletedPet.Id != secondPet.Id {
		t.Fatalf("pet.delete id = %q, want %q", deletedPet.Id, secondPet.Id)
	}
}
