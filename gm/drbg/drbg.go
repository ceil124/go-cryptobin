package drbg

import (
    "hash"
    "errors"
)

type Drbg struct {
    digest hash.Hash
    v [HASH_DRBG_MAX_SEED_SIZE]byte
    c [HASH_DRBG_MAX_SEED_SIZE]byte
    seedlen int
    reseedCounter uint64
    isGm bool
}

func New(digest hash.Hash, isGm bool, entropy []byte, nonce []byte, personalstr []byte) *Drbg {
    drbg := new(Drbg)
    drbg.init(digest, entropy, nonce, personalstr)

    /* set isGm */
    drbg.isGm = isGm

    return drbg
}

func (this *Drbg) init(digest hash.Hash, entropy []byte, nonce []byte, personalstr []byte) {
    var seedMaterial []byte
    var seedMaterialLen int
    var buf [1 + HASH_DRBG_MAX_SEED_SIZE]byte

    entropyLen := len(entropy)
    nonceLen := len(nonce)
    personalstrLen := len(personalstr)

    /* set digest */
    this.digest = digest

    /* set seedlen */
    if digest.Size() <= 32 {
        this.seedlen = HASH_DRBG_SM3_SEED_SIZE
    } else {
        this.seedlen = HASH_DRBG_SHA512_SEED_SIZE
    }

    /* seedMaterial = entropy_input || nonce || personalization_string */
    seedMaterialLen = entropyLen + nonceLen + personalstrLen
    seedMaterial = make([]byte, seedMaterialLen)

    copy(seedMaterial, entropy)
    copy(seedMaterial[entropyLen:], nonce)
    copy(seedMaterial[entropyLen+nonceLen:], personalstr)

    /* V = Hash_df (seedMaterial, seedlen) */
    HashDF(this.digest, seedMaterial, this.v[:this.seedlen])

    /* C = Hash_df ((0x00 || V), seedlen) */
    buf[0] = 0x00
    copy(buf[1:], this.v[:this.seedlen])
    HashDF(this.digest, buf[:1 + this.seedlen], this.c[:this.seedlen])

    /* reseedCounter = 1 */
    this.reseedCounter = 1
}

func (this *Drbg) Reseed(entropy []byte, additional []byte) {
    var seedMaterial []byte
    var seedMaterialLen int
    var buf [1 + HASH_DRBG_MAX_SEED_SIZE]byte

    entropyLen := len(entropy)
    additionalLen := len(additional)

    /* seedMaterial = 0x01 || V || entropy_input || additional_input */
    seedMaterialLen = 1 + this.seedlen + entropyLen + additionalLen
    seedMaterial = make([]byte, seedMaterialLen)

    seedMaterial[0] = 0x01
    if this.isGm {
        copy(seedMaterial[1:], entropy)
        copy(seedMaterial[1+entropyLen:], this.v[:this.seedlen])
    } else {
        copy(seedMaterial[1:], this.v[:this.seedlen])
        copy(seedMaterial[1+this.seedlen:], entropy)
    }

    copy(seedMaterial[1+this.seedlen+entropyLen:], additional)

    /* V = Hash_df(seedMaterial, seedlen) */
    HashDF(this.digest, seedMaterial, this.v[:this.seedlen])

    /* C = Hash_df((0x00 || V), seedlen) */
    buf[0] = 0x00
    copy(buf[1:], this.v[:this.seedlen])
    HashDF(this.digest, buf[:1 + this.seedlen], this.c[:this.seedlen])

    /* reseedCounter = 1 */
    this.reseedCounter = 1
}

func (this *Drbg) hashgen(out []byte) {
    var h hash.Hash
    var data [HASH_DRBG_MAX_SEED_SIZE]byte
    var dgst []byte = make([]byte, DIGEST_MAX_SIZE)
    var length int

    h = this.digest

    if this.isGm {
        h.Reset()
        h.Write(this.v[:this.seedlen])
        dgst = h.Sum(nil)

        copy(out[:], dgst)
    } else {
        /* data = V */
        copy(data[:], this.v[:this.seedlen])

        outlen := len(out)

        var nlength int = 0
        for outlen > 0 {
            /* output Hash(data) */
            h.Reset()
            h.Write(data[:this.seedlen])
            dgst = h.Sum(nil)

            length = len(dgst)
            if outlen < length {
                length = outlen
            }

            copy(out[nlength:], dgst[:length])

            outlen -= length
            nlength += length

            /* data = (data + 1) mod 2^seedlen */
            drbg_add1(data[:], this.seedlen)
        }
    }
}

func (this *Drbg) Generate(out []byte, additional []byte) error {
    var ctx hash.Hash
    var prefix byte
    var T [HASH_DRBG_MAX_SEED_SIZE]byte
    var dgst []byte = make([]byte, DIGEST_MAX_SIZE)

    if this.reseedCounter > HASH_DRBG_RESEED_INTERVAL {
        return errors.New("drbg: reseed counter error")
    }

    ctx = this.digest

    if len(additional) > 0 {
        /* w = Hash (0x02 || V || additional_input) */
        prefix = 0x02

        ctx.Reset()
        ctx.Write([]byte{prefix})
        ctx.Write(this.v[:this.seedlen])
        ctx.Write(additional)
        dgst = ctx.Sum(nil)

        /* V = (V + w) mod 2^seedlen */
        T = [HASH_DRBG_MAX_SEED_SIZE]byte{}
        copy(T[this.seedlen - len(dgst):], dgst[:])
        drbg_add(this.v[:], T[:], this.seedlen)
    }

    /* (returned_bits) = Hashgen (requested_number_of_bits, V). */
    this.hashgen(out)

    /* H = Hash (0x03 || V). */
    prefix = 0x03
    ctx.Reset()
    ctx.Write([]byte{prefix})
    ctx.Write(this.v[:this.seedlen])
    dgst = ctx.Sum(nil)

    /* V = (V + H + C + reseedCounter) mod 2^seedlen */
    T = [HASH_DRBG_MAX_SEED_SIZE]byte{}
    copy(T[this.seedlen - len(dgst):], dgst[:])
    drbg_add(this.v[:], T[:], this.seedlen)

    drbg_add(this.v[:], this.c[:], this.seedlen)

    T = [HASH_DRBG_MAX_SEED_SIZE]byte{}
    PUTU64(T[this.seedlen - 8:], this.reseedCounter)
    drbg_add(this.v[:], T[:], this.seedlen)

    /* reseedCounter = reseedCounter + 1 */
    this.reseedCounter++

    return nil
}