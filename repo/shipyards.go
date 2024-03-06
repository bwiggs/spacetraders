package repo

import (
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/pkg/errors"
)

func (r *Repo) UpsertShipyard(shipyard api.Shipyard) (err error) {
	// more details
	if shipyard.Ships != nil {
		err = r.UpsertShips(shipyard.Ships)
		if err != nil {
			return errors.Wrap(err, "UpsertShips")
		}

		err = r.UpsertShipyardShips(shipyard.Symbol, shipyard.Ships)
		if err != nil {
			return errors.Wrap(err, "UpsertShipyardShips")
		}

	} else {
		err = r.UpsertShipTypes(shipyard.ShipTypes)
		if err != nil {
			return errors.Wrap(err, "UpsertShipTypes")
		}

		err = r.UpsertShipyardShipType(shipyard.Symbol, shipyard.ShipTypes)
		if err != nil {
			return errors.Wrap(err, "UpsertShipyardShipType")
		}
	}

	return nil
}

func (r *Repo) UpsertShipTypes(shipTypes []api.ShipyardShipTypesItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	upsertShip, err := tx.Prepare("INSERT OR REPLACE INTO ships (type) values (?)")
	if err != nil {
		return err
	}
	defer upsertShip.Close()

	for _, st := range shipTypes {
		_, err = upsertShip.Exec(st.Type)
		if err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}

func (r *Repo) UpsertShips(ships []api.ShipyardShip) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	upsertShip, err := tx.Prepare("INSERT OR REPLACE INTO ships (type, name, description) values (?, ?, ?)")
	if err != nil {
		return err
	}
	defer upsertShip.Close()

	for _, ship := range ships {
		_, err = upsertShip.Exec(ship.Type, ship.Name, ship.Description)
		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}

func (r *Repo) UpsertShipyardShips(waypoint string, ships []api.ShipyardShip) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	upsert, err := tx.Prepare("INSERT OR REPLACE INTO shipyards (waypoint, ship, supply, bid) values (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer upsert.Close()

	for _, ship := range ships {
		_, err = upsert.Exec(waypoint, ship.Type, ship.Supply, ship.PurchasePrice)
		if err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}

func (r *Repo) UpsertShipyardShipType(waypoint string, ships []api.ShipyardShipTypesItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	upsert, err := tx.Prepare("INSERT OR REPLACE INTO shipyards (waypoint, ship) values (?, ?)")
	if err != nil {
		return err
	}
	defer upsert.Close()

	for _, ship := range ships {
		_, err = upsert.Exec(waypoint, ship.Type)
		if err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}
