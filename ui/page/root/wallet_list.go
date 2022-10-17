package root

import (
	"gioui.org/layout"

	"github.com/decred/dcrd/dcrutil/v4"
	"gitlab.com/raedah/cryptopower/libwallet/assets/dcr"
	"gitlab.com/raedah/cryptopower/listeners"
	"gitlab.com/raedah/cryptopower/ui/cryptomaterial"
	"gitlab.com/raedah/cryptopower/ui/load"
	"gitlab.com/raedah/cryptopower/ui/modal"
	"gitlab.com/raedah/cryptopower/ui/values"
	"gitlab.com/raedah/cryptopower/wallet"
)

func (pg *WalletDexServerSelector) initWalletSelectorOptions() {
	pg.walletsList = pg.Theme.NewClickableList(layout.Vertical)
	pg.BTCwalletsList = pg.Theme.NewClickableList(layout.Vertical)
	pg.watchOnlyWalletsList = pg.Theme.NewClickableList(layout.Vertical)
}

func (pg *WalletDexServerSelector) loadWallets() {
	wallets := pg.WL.SortedWalletList()
	mainWalletList := make([]*load.WalletItem, 0)
	watchOnlyWalletList := make([]*load.WalletItem, 0)

	for _, wal := range wallets {
		accountsResult, err := wal.GetAccountsRaw()
		if err != nil {
			continue
		}

		var totalBalance int64
		for _, acc := range accountsResult.Acc {
			totalBalance += acc.TotalBalance
		}

		// sort wallets into normal wallet and watchonly wallets
		if wal.IsWatchingOnlyWallet() {
			listItem := &load.WalletItem{
				Wallet:       wal,
				TotalBalance: dcrutil.Amount(totalBalance).String(),
			}

			watchOnlyWalletList = append(watchOnlyWalletList, listItem)
		} else {
			listItem := &load.WalletItem{
				Wallet:       wal,
				TotalBalance: dcrutil.Amount(totalBalance).String(),
			}

			mainWalletList = append(mainWalletList, listItem)
		}
	}

	pg.listLock.Lock()
	pg.mainWalletList = mainWalletList
	pg.watchOnlyWalletList = watchOnlyWalletList
	pg.listLock.Unlock()

	pg.loadBadWallets()
}

func (pg *WalletDexServerSelector) loadBTCWallets() {
	wallets := pg.WL.SortedBTCWalletList()
	mainWalletList := make([]*load.BTCWalletItem, 0)
	watchOnlyWalletList := make([]*load.BTCWalletItem, 0)

	for _, wal := range wallets {
		accountsResult, err := wal.GetAccountsRaw()
		if err != nil {
			continue
		}

		var totalBalance int64
		for _, acc := range accountsResult.Accounts {
			totalBalance += int64(acc.TotalBalance)
		}

		// sort wallets into normal wallet and watchonly wallets
		if wal.IsWatchingOnlyWallet() {
			listItem := &load.BTCWalletItem{
				Wallet:       wal,
				TotalBalance: dcrutil.Amount(totalBalance).String(),
			}

			watchOnlyWalletList = append(watchOnlyWalletList, listItem)
		} else {
			listItem := &load.BTCWalletItem{
				Wallet:       wal,
				TotalBalance: dcrutil.Amount(totalBalance).String(),
			}

			mainWalletList = append(mainWalletList, listItem)
		}
	}

	pg.listLock.Lock()
	pg.mainBTCWalletList = mainWalletList
	pg.listLock.Unlock()

	pg.loadBadWallets()
}

func (pg *WalletDexServerSelector) loadBadWallets() {
	badWallets := pg.WL.MultiWallet.BadWallets()
	pg.badWalletsList = make([]*badWalletListItem, 0, len(badWallets))
	for _, badWallet := range badWallets {
		listItem := &badWalletListItem{
			DCRAsset:  badWallet,
			deleteBtn: pg.Theme.OutlineButton(values.String(values.StrDeleted)),
		}
		listItem.deleteBtn.Color = pg.Theme.Color.Danger
		listItem.deleteBtn.Inset = layout.Inset{}
		pg.badWalletsList = append(pg.badWalletsList, listItem)
	}
}

func (pg *WalletDexServerSelector) deleteBadWallet(badWalletID int) {
	warningModal := modal.NewCustomModal(pg.Load).
		Title(values.String(values.StrRemoveWallet)).
		Body(values.String(values.StrWalletRestoreMsg)).
		SetNegativeButtonText(values.String(values.StrCancel)).
		PositiveButtonStyle(pg.Load.Theme.Color.Surface, pg.Load.Theme.Color.Danger).
		SetPositiveButtonText(values.String(values.StrRemove)).
		SetPositiveButtonCallback(func(_ bool, im *modal.InfoModal) bool {
			err := pg.WL.MultiWallet.DeleteBadDCRWallet(badWalletID)
			if err != nil {
				errorModal := modal.NewErrorModal(pg.Load, err.Error(), modal.DefaultClickFunc())
				pg.ParentWindow().ShowModal(errorModal)
				return false
			}
			infoModal := modal.NewSuccessModal(pg.Load, values.String(values.StrWalletRemoved), modal.DefaultClickFunc())
			pg.ParentWindow().ShowModal(infoModal)
			pg.loadBadWallets() // refresh bad wallets list
			pg.ParentWindow().Reload()
			return true
		})
	pg.ParentWindow().ShowModal(warningModal)
}

func (pg *WalletDexServerSelector) syncStatusIcon(gtx C, wallet *dcr.DCRAsset) D {
	var (
		syncStatusIcon *cryptomaterial.Image
		syncStatus     string
	)

	switch {
	case wallet.IsSynced():
		syncStatusIcon = pg.Theme.Icons.SuccessIcon
		syncStatus = values.String(values.StrSynced)
	case wallet.IsSyncing():
		syncStatusIcon = pg.Theme.Icons.SyncingIcon
		syncStatus = values.String(values.StrSyncingState)
	default:
		syncStatusIcon = pg.Theme.Icons.NotSynced
		syncStatus = values.String(values.StrWalletNotSynced)
	}

	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(syncStatusIcon.Layout16dp),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Left: values.MarginPadding5,
			}.Layout(gtx, pg.Theme.Label(values.TextSize16, syncStatus).Layout)
		}),
	)
}

func (pg *WalletDexServerSelector) walletListLayout(gtx C) D {
	walletSections := []func(gtx C) D{
		pg.walletList,
	}

	if len(pg.watchOnlyWalletList) != 0 {
		walletSections = append(walletSections, pg.watchOnlyWalletSection)
	}

	if len(pg.badWalletsList) != 0 {
		walletSections = append(walletSections, pg.badWalletsSection)
	}
	list := &layout.List{
		Axis: layout.Vertical,
	}
	return list.Layout(gtx, len(walletSections), func(gtx C, i int) D {
		return walletSections[i](gtx)
	})
}

func (pg *WalletDexServerSelector) BTCwalletListLayout(gtx C) D {
	walletSections := []func(gtx C) D{
		pg.BTCwalletList,
	}

	list := &layout.List{
		Axis: layout.Vertical,
	}
	return list.Layout(gtx, len(walletSections), func(gtx C, i int) D {
		return walletSections[i](gtx)
	})
}

func (pg *WalletDexServerSelector) walletList(gtx C) D {
	pg.listLock.Lock()
	mainWalletList := pg.mainWalletList
	pg.listLock.Unlock()

	return pg.walletsList.Layout(gtx, len(mainWalletList), func(gtx C, i int) D {
		return pg.walletWrapper(gtx, mainWalletList[i], false)
	})
}

func (pg *WalletDexServerSelector) BTCwalletList(gtx C) D {
	pg.listLock.Lock()
	mainWalletList := pg.mainBTCWalletList
	pg.listLock.Unlock()

	return pg.BTCwalletsList.Layout(gtx, len(mainWalletList), func(gtx C, i int) D {
		return pg.BTCwalletWrapper(gtx, mainWalletList[i], false)
	})
}

func (pg *WalletDexServerSelector) watchOnlyWalletSection(gtx C) D {
	pg.listLock.Lock()
	watchOnlyWalletList := pg.watchOnlyWalletList
	pg.listLock.Unlock()

	return pg.watchOnlyWalletsList.Layout(gtx, len(watchOnlyWalletList), func(gtx C, i int) D {
		return pg.walletWrapper(gtx, watchOnlyWalletList[i], true)
	})
}

func (pg *WalletDexServerSelector) badWalletsSection(gtx C) D {
	m20 := values.MarginPadding20
	m10 := values.MarginPadding10

	layoutBadWallet := func(gtx C, badWallet *badWalletListItem, lastItem bool) D {
		return layout.Inset{Top: m10, Bottom: m10}.Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Flex{}.Layout(gtx,
						layout.Rigid(pg.Theme.Body2(badWallet.Name).Layout),
						layout.Flexed(1, func(gtx C) D {
							return layout.E.Layout(gtx, func(gtx C) D {
								return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, badWallet.deleteBtn.Layout)
							})
						}),
					)
				}),
				layout.Rigid(func(gtx C) D {
					if lastItem {
						return D{}
					}
					return layout.Inset{Top: values.MarginPadding10, Left: values.MarginPadding38, Right: values.MarginPaddingMinus10}.Layout(gtx, func(gtx C) D {
						return pg.Theme.Separator().Layout(gtx)
					})
				}),
			)
		})
	}

	card := pg.Theme.Card()
	card.Color = pg.Theme.Color.Surface
	card.Radius = cryptomaterial.Radius(10)

	sectionTitleLabel := pg.Theme.Body1("Bad Wallets") // TODO: localize string
	sectionTitleLabel.Color = pg.Theme.Color.GrayText2

	return card.Layout(gtx, func(gtx C) D {
		return layout.Inset{Top: m20, Left: m20}.Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(sectionTitleLabel.Layout),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: m10, Bottom: m10}.Layout(gtx, pg.Theme.Separator().Layout)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Right: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
						return pg.Theme.NewClickableList(layout.Vertical).Layout(gtx, len(pg.badWalletsList), func(gtx C, i int) D {
							return layoutBadWallet(gtx, pg.badWalletsList[i], i == len(pg.badWalletsList)-1)
						})
					})
				}),
			)
		})
	})
}

func (pg *WalletDexServerSelector) walletWrapper(gtx C, item *load.WalletItem, isWatchingOnlyWallet bool) D {
	pg.shadowBox.SetShadowRadius(14)
	return cryptomaterial.LinearLayout{
		Width:      cryptomaterial.WrapContent,
		Height:     cryptomaterial.WrapContent,
		Padding:    layout.UniformInset(values.MarginPadding9),
		Background: pg.Theme.Color.Surface,
		Alignment:  layout.Middle,
		Shadow:     pg.shadowBox,
		Margin:     layout.UniformInset(values.MarginPadding5),
		Border:     cryptomaterial.Border{Radius: cryptomaterial.Radius(14)},
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Right: values.MarginPadding10,
				Left:  values.MarginPadding10,
			}.Layout(gtx, func(gtx C) D {
				if isWatchingOnlyWallet {
					return pg.Theme.Icons.DcrWatchOnly.Layout36dp(gtx)
				}
				return pg.Theme.Icons.DecredSymbol2.LayoutSize(gtx, values.MarginPadding30)
			})
		}),
		layout.Rigid(pg.Theme.Label(values.TextSize16, item.Wallet.Name).Layout),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						if len(item.Wallet.EncryptedSeed) > 0 {
							return layout.Flex{
								Axis:      layout.Horizontal,
								Alignment: layout.Middle,
							}.Layout(gtx,
								layout.Rigid(pg.Theme.Icons.RedAlert.Layout16dp),
								layout.Rigid(func(gtx C) D {
									return layout.Inset{
										Right: values.MarginPadding10,
									}.Layout(gtx, pg.Theme.Label(values.TextSize16, values.String(values.StrNotBackedUp)).Layout)
								}),
							)
						}
						return D{}
					}),
					layout.Rigid(func(gtx C) D {
						return pg.syncStatusIcon(gtx, item.Wallet)
					}),
				)
			})
		}),
	)
}

func (pg *WalletDexServerSelector) BTCwalletWrapper(gtx C, item *load.BTCWalletItem, isWatchingOnlyWallet bool) D {
	pg.shadowBox.SetShadowRadius(14)
	return cryptomaterial.LinearLayout{
		Width:      cryptomaterial.WrapContent,
		Height:     cryptomaterial.WrapContent,
		Padding:    layout.UniformInset(values.MarginPadding9),
		Background: pg.Theme.Color.Surface,
		Alignment:  layout.Middle,
		Shadow:     pg.shadowBox,
		Margin:     layout.UniformInset(values.MarginPadding5),
		Border:     cryptomaterial.Border{Radius: cryptomaterial.Radius(14)},
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Right: values.MarginPadding10,
				Left:  values.MarginPadding10,
			}.Layout(gtx, func(gtx C) D {
				if isWatchingOnlyWallet {
					return pg.Theme.Icons.DcrWatchOnly.Layout36dp(gtx)
				}
				return pg.Theme.Icons.BTC.LayoutSize(gtx, values.MarginPadding30)
			})
		}),
		layout.Rigid(pg.Theme.Label(values.TextSize16, item.Wallet.Name).Layout),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx)
			})
		}),
	)
}

// start sync listener
func (pg *WalletDexServerSelector) listenForNotifications() {
	if pg.isListenerAdded {
		return
	}

	pg.isListenerAdded = true

	for k, w := range pg.WL.SortedWalletList() {
		syncListener := listeners.NewSyncProgress()
		err := w.AddSyncProgressListener(syncListener, WalletDexServerSelectorID)
		if err != nil {
			log.Errorf("Error adding sync progress listener: %v", err)
			return
		}

		go func(wal *dcr.DCRAsset, k int) {
			for {
				select {
				case n := <-syncListener.SyncStatusChan:
					if n.Stage == wallet.SyncCompleted {
						pg.ParentWindow().Reload()
					}
				case <-pg.ctx.Done():
					wal.RemoveSyncProgressListener(WalletDexServerSelectorID)
					close(syncListener.SyncStatusChan)
					syncListener = nil
					pg.isListenerAdded = false
					return
				}
			}
		}(w, k)
	}
}
