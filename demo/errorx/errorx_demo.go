package main

import (
	"log"

	"github.com/madlabx/pkgx/errors"
)

func f1() error {
	err := errors.New("New")
	log.Printf("f1:%+v\n", err)
	return err
}

func f2() error {
	err := f1()
	log.Printf("f2:%+v\n", err)
	return errors.WithStack(err)
}

func f3() error {
	err := f2()
	log.Printf("f3:%+v\n", err)
	return errors.WithMessage(err, "WithMessage")
}

func main() {
	err := f2()
	cause := errors.Cause(err)
	log.Printf("errors.Cause=%#v\n", cause)

	unwrap := errors.Unwrap(err)

	log.Printf("errors.Unwrap=%#v\n", unwrap)

	log.Printf("f4:%+v\n", err)
	err1 := errors.New("New2")
	log.Printf("erro1:%#v", err1)
	err2 := errors.Wrap(err1, "New2+p")
	log.Printf("erro2:%#v", err2)

	err3 := errors.Wrapf(err2, "New2+q")
	log.Printf("erro3:%#v", err3)

	err4 := errors.Unwrap(err3)
	log.Printf("erro3:%#v", err4)

	err5 := errors.Cause(err3)
	log.Printf("err5:%#v", err5)

}
